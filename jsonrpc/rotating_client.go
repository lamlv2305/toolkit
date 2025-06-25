package jsonrpc

import (
	"context"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
)

// ClientType represents the type of RPC client
type ClientType int

const (
	HTTPClient ClientType = iota
	WSClient
)

// RotatingClient wraps ethclient.Client with endpoint information and automatic failure detection
type RotatingClient struct {
	*ethclient.Client
	rpcClient  *rpc.Client
	endpoint   string
	rotator    *Rotator
	index      int
	clientType ClientType
}

// Override key methods to add automatic failure detection

func (c *RotatingClient) CallContext(ctx context.Context, result any, method string, args ...any) error {
	err := c.rpcClient.CallContext(ctx, result, method, args...)
	if err != nil {
		// whatever error, keep it as failed
		c.rotator.markFailed(c.index, err)
	}
	return err
}

func (c *RotatingClient) BatchCallContext(ctx context.Context, batch []rpc.BatchElem) error {
	err := c.rpcClient.BatchCallContext(ctx, batch)
	if err != nil {
		// whatever error, keep it as failed
		c.rotator.markFailed(c.index, err)
		return err
	}

	return nil
}
