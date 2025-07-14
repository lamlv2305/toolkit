package jsonrpc

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
)

type RoratorOption func(*Rotator)

func WithNotifier(notifier RPCHealthNotifier) RoratorOption {
	return func(r *Rotator) {
		r.notifier = notifier
	}
}

// RPCHealthNotifier interface for notifying when RPCs fail or recover
type RPCHealthNotifier interface {
	NotifyRPCFailure(endpoint string, err error)
	NotifyRPCRecovery(endpoint string)
}

// Rotator manages a pool of JSON-RPC clients with automatic failover and recovery
type Rotator struct {
	nativeSymbol  string
	nativeDecimal int64
	clients       []*RotatingClient
	activeClients map[int]bool
	excludedUntil map[int]time.Time
	mutex         sync.RWMutex
	notifier      RPCHealthNotifier
	currentIndex  int
}

func NewDefaultJsonrpcRotator(ctx context.Context, chainId int64, options ...RoratorOption) (*Rotator, error) {
	if err := ensureDataLoaded(ctx); err != nil {
		return nil, err
	}

	mu.RLock()
	chainData, exists := chains[chainId]
	mu.RUnlock()

	if !exists {
		return nil, errors.New("no chain data available for the specified chain ID")
	}

	decimal := chainData.NativeDecimal
	if decimal == 0 {
		decimal = 18 // Default to 18 decimals if not specified
	}

	if len(chainData.PublicRPCs) == 0 {
		return nil, errors.New("no public RPC endpoints available for the specified chain ID")
	}

	rotator, err := NewJsonrpcRotator(chainData.PublicRPCs, chainData.NativeSymbol, decimal, nil)
	if err != nil {
		return nil, err
	}

	for _, option := range options {
		option(rotator)
	}

	return rotator, nil
}

// NewJsonrpcRotator creates a new JsonrpcRotator with the given endpoints
func NewJsonrpcRotator(endpoints []string, symbol string, decimal int64, notifier RPCHealthNotifier) (*Rotator, error) {
	if len(endpoints) == 0 {
		return nil, errors.New("no RPC endpoints provided")
	}

	rotator := &Rotator{
		nativeSymbol:  symbol,
		nativeDecimal: decimal,
		clients:       make([]*RotatingClient, 0, len(endpoints)),
		activeClients: make(map[int]bool),
		excludedUntil: make(map[int]time.Time),
		notifier:      notifier,
		currentIndex:  0,
	}

	for i, url := range endpoints {
		clientType := HTTPClient
		if strings.HasPrefix(url, "ws://") || strings.HasPrefix(url, "wss://") {
			clientType = WSClient
		}

		rpcClient, err := rpc.DialContext(context.Background(), url)
		if err != nil {
			continue // Skip this endpoint but continue with others
		}

		client := ethclient.NewClient(rpcClient)

		rc := &RotatingClient{
			Client:     client,
			rpcClient:  rpcClient,
			endpoint:   url,
			rotator:    rotator,
			index:      i,
			clientType: clientType,
		}

		rotator.clients = append(rotator.clients, rc)
		rotator.activeClients[i] = true
	}

	if len(rotator.clients) == 0 {
		return nil, errors.New("failed to connect to any RPC endpoints")
	}

	return rotator, nil
}

// GetClient returns the next available RPC client
func (r *Rotator) GetSymbol() string {
	return r.nativeSymbol
}

func (r *Rotator) GetDecimal() int64 {
	return r.nativeDecimal
}

// GetClient returns the next available RPC client
func (r *Rotator) GetClient() (*RotatingClient, error) {
	return r.getClientByType(HTTPClient)
}

// GetWSClient returns the next available WebSocket RPC client
func (r *Rotator) GetWSClient() (*RotatingClient, error) {
	return r.getClientByType(WSClient)
}

// getClientByType returns the next available client of the specified type
func (r *Rotator) getClientByType(clientType ClientType) (*RotatingClient, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	// Check for expired exclusions and restore them
	now := time.Now()
	for idx, until := range r.excludedUntil {
		if now.After(until) {
			delete(r.excludedUntil, idx)
			r.activeClients[idx] = true
			if r.notifier != nil {
				r.notifier.NotifyRPCRecovery(r.clients[idx].endpoint)
			}
		}
	}

	// Find an active client of the specified type using round-robin
	startIndex := r.currentIndex
	checkedIndexes := 0

	for checkedIndexes < len(r.clients) {
		idx := (startIndex + checkedIndexes) % len(r.clients)
		client := r.clients[idx]

		if r.activeClients[idx] && client.clientType == clientType {
			r.currentIndex = (idx + 1) % len(r.clients) // Move to next for next time
			return client, nil
		}

		checkedIndexes++
	}

	if clientType == WSClient {
		return nil, errors.New("no active WebSocket RPC endpoints available")
	}

	return nil, errors.New("no active HTTP RPC endpoints available")
}

// markFailed marks an endpoint as failed and excludes it temporarily
// This is now internal and called automatically by the RotatingClient
func (r *Rotator) markFailed(index int, err error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.activeClients[index] = false
	r.excludedUntil[index] = time.Now().Add(15 * time.Second)

	if r.notifier != nil {
		r.notifier.NotifyRPCFailure(r.clients[index].endpoint, err)
	}
}

// Close closes all client connections
func (r *Rotator) Close() {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	for _, client := range r.clients {
		client.Close()
	}
}
