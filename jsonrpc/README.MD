# JSON-RPC Rotator

A robust JSON-RPC client rotator for Ethereum and other EVM-compatible blockchains with automatic failover, recovery, and health monitoring.

## Features

- **Automatic Failover**: Seamlessly switches between RPC endpoints when failures occur
- **Health Recovery**: Automatically re-enables failed endpoints after recovery periods
- **Chain Data Integration**: Loads public RPC endpoints from chainlist.org
- **WebSocket Support**: Handles both HTTP and WebSocket RPC connections
- **Thread-Safe**: Concurrent access with proper mutex protection
- **Configurable Options**: Flexible configuration with functional options pattern
- **Batch Requests**: Support for RPC batch operations
- **Health Notifications**: Optional callback system for monitoring RPC health

## Installation

```bash
go get github.com/lamlv2305/rok/jsonrpc
```

## Quick Start

### Using Default Chainlist RPCs

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/lamlv2305/rok/jsonrpc"
)

func main() {
    ctx := context.Background()

    // Create rotator for Ethereum mainnet (chain ID 1)
    rotator, err := jsonrpc.NewDefaultJsonrpcRotator(ctx, 1)
    if err != nil {
        log.Fatal(err)
    }
    defer rotator.Close()

    // Get a client and use it
    client, err := rotator.GetClient()
    if err != nil {
        log.Fatal(err)
    }

    blockNumber, err := client.BlockNumber(ctx)
    if err != nil {
        log.Printf("Error: %v", err)
        return
    }

    fmt.Printf("Latest block: %d\n", blockNumber)
}
```

### Using Custom RPC Endpoints

```go
package main

import (
    "context"
    "log"

    "github.com/lamlv2305/rok/jsonrpc"
)

func main() {
    // Define your RPC endpoints
    endpoints := []string{
        "https://mainnet.infura.io/v3/YOUR_API_KEY",
        "https://eth-mainnet.alchemyapi.io/v2/YOUR_API_KEY",
        "https://cloudflare-eth.com",
        "https://rpc.ankr.com/eth",
    }

    // Create rotator with custom endpoints
    rotator, err := jsonrpc.NewJsonrpcRotator(endpoints, "ETH", 18, nil)
    if err != nil {
        log.Fatal(err)
    }
    defer rotator.Close()

    // Use the rotator...
}
```

## Configuration Options

### Health Monitoring

Implement the `RPCHealthNotifier` interface to receive health notifications:

```go
type MyNotifier struct{}

func (n *MyNotifier) NotifyRPCFailure(endpoint string, err error) {
    log.Printf("RPC failed: %s - %v", endpoint, err)
}

func (n *MyNotifier) NotifyRPCRecovery(endpoint string) {
    log.Printf("RPC recovered: %s", endpoint)
}

// Use with default rotator
notifier := &MyNotifier{}
rotator, err := jsonrpc.NewDefaultJsonrpcRotator(ctx, 1,
    jsonrpc.WithNotifier(notifier))
```

### WebSocket Support

```go
// Mix HTTP and WebSocket endpoints
endpoints := []string{
    "https://mainnet.infura.io/v3/YOUR_API_KEY",     // HTTP
    "wss://mainnet.infura.io/ws/v3/YOUR_API_KEY",   // WebSocket
    "https://cloudflare-eth.com",                    // HTTP
}

rotator, err := jsonrpc.NewJsonrpcRotator(endpoints, "ETH", 18, nil)
if err != nil {
    log.Fatal(err)
}

// Get WebSocket client specifically
wsClient, err := rotator.GetWSClient()
if err != nil {
    // Fallback to HTTP client
    client, err := rotator.GetClient()
    // ...
}
```

## Supported Chains

The rotator automatically loads chain data from chainlist.org, supporting hundreds of EVM-compatible chains:

- **Ethereum Mainnet** (Chain ID: 1)
- **Polygon** (Chain ID: 137)
- **Binance Smart Chain** (Chain ID: 56)
- **Arbitrum One** (Chain ID: 42161)
- **Optimism** (Chain ID: 10)
- **Avalanche C-Chain** (Chain ID: 43114)
- And many more...

### Getting Chain Information

```go
ctx := context.Background()

// Get chain data
chainData, err := jsonrpc.GetChainData(ctx, 1) // Ethereum
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Chain: %s\n", chainData.ChainName)
fmt.Printf("Symbol: %s\n", chainData.NativeSymbol)
fmt.Printf("Decimals: %d\n", chainData.NativeDecimal)
fmt.Printf("RPCs: %v\n", chainData.PublicRPCs)

// Get all supported chains
allChains, err := jsonrpc.GetAllChains(ctx)
if err != nil {
    log.Fatal(err)
}

for chainID, data := range allChains {
    fmt.Printf("Chain %d: %s (%s)\n", chainID, data.ChainName, data.NativeSymbol)
}
```

## Advanced Usage

### Retry Logic with Failover

```go
func makeRequestWithRetry(rotator *jsonrpc.Rotator) error {
    ctx := context.Background()
    maxRetries := 3

    for attempt := 0; attempt < maxRetries; attempt++ {
        client, err := rotator.GetClient()
        if err != nil {
            time.Sleep(time.Second * time.Duration(attempt+1))
            continue
        }

        blockNumber, err := client.BlockNumber(ctx)
        if err != nil {
            // Rotator automatically marks failed endpoints
            time.Sleep(time.Second * time.Duration(attempt+1))
            continue
        }

        fmt.Printf("Success! Block: %d\n", blockNumber)
        return nil
    }

    return fmt.Errorf("all retry attempts failed")
}
```

### Batch Requests

```go
import "github.com/ethereum/go-ethereum/rpc"

func performBatchRequest(rotator *jsonrpc.Rotator) error {
    client, err := rotator.GetClient()
    if err != nil {
        return err
    }

    ctx := context.Background()

    var blockNumber *big.Int
    var gasPrice *big.Int

    batch := []rpc.BatchElem{
        {
            Method: "eth_blockNumber",
            Result: &blockNumber,
        },
        {
            Method: "eth_gasPrice",
            Result: &gasPrice,
        },
    }

    err = client.BatchCallContext(ctx, batch)
    if err != nil {
        return err
    }

    fmt.Printf("Block: %v, Gas Price: %v\n", blockNumber, gasPrice)
    return nil
}
```

### Multiple Chain Support

```go
func useMultipleChains() {
    ctx := context.Background()
    notifier := &MyNotifier{}

    // Create rotators for different chains
    chains := map[string]int64{
        "ethereum": 1,
        "polygon":  137,
        "bsc":      56,
    }

    rotators := make(map[string]*jsonrpc.Rotator)

    for name, chainID := range chains {
        rotator, err := jsonrpc.NewDefaultJsonrpcRotator(ctx, chainID,
            jsonrpc.WithNotifier(notifier))
        if err != nil {
            log.Printf("Failed to create %s rotator: %v", name, err)
            continue
        }

        rotators[name] = rotator
        defer rotator.Close()
    }

    // Use different chains
    if ethRotator, exists := rotators["ethereum"]; exists {
        client, _ := ethRotator.GetClient()
        // Perform Ethereum operations...
    }

    if polyRotator, exists := rotators["polygon"]; exists {
        client, _ := polyRotator.GetClient()
        // Perform Polygon operations...
    }
}
```

## API Reference

### Types

#### `Rotator`

Main struct that manages RPC client rotation.

**Methods:**

- `GetClient() (*RotatingClient, error)` - Get next available HTTP client
- `GetWSClient() (*RotatingClient, error)` - Get next available WebSocket client
- `GetSymbol() string` - Get native token symbol
- `GetDecimal() int64` - Get native token decimals
- `Close()` - Close all client connections

#### `RotatingClient`

Wraps `ethclient.Client` with automatic failure detection.

**Methods:**

- All `ethclient.Client` methods
- `CallContext(ctx, result, method, args...)` - Raw RPC call with failure detection
- `BatchCallContext(ctx, batch)` - Batch RPC calls with failure detection

#### `RPCHealthNotifier`

Interface for receiving RPC health notifications.

**Methods:**

- `NotifyRPCFailure(endpoint string, err error)`
- `NotifyRPCRecovery(endpoint string)`

### Functions

#### Constructor Functions

```go
// Create rotator using chainlist data
func NewDefaultJsonrpcRotator(ctx context.Context, chainID int64, options ...RoratorOption) (*Rotator, error)

// Create rotator with custom endpoints
func NewJsonrpcRotator(endpoints []string, symbol string, decimal int64, notifier RPCHealthNotifier) (*Rotator, error)
```

#### Option Functions

```go
// Set health notifier
func WithNotifier(notifier RPCHealthNotifier) RoratorOption
```

#### Chain Data Functions

```go
// Get public RPCs for a chain
func GetPublicRPCs(ctx context.Context, chainID int64) ([]string, error)

// Get native token symbol
func GetNativeSymbol(ctx context.Context, chainID int64) (string, error)

// Get chain name
func GetChainName(ctx context.Context, chainID int64) (string, error)

// Get complete chain data
func GetChainData(ctx context.Context, chainID int64) (*ChainData, error)

// Get all supported chains
func GetAllChains(ctx context.Context) (map[int64]*ChainData, error)
```

## Error Handling

The rotator automatically handles common RPC errors:

- **Connection timeouts**: Endpoint marked as failed, rotates to next
- **Network errors**: Automatic failover to healthy endpoints
- **Invalid responses**: Failed endpoint excluded temporarily
- **Rate limiting**: Rotates to different endpoint to distribute load

Failed endpoints are automatically re-enabled after 15 seconds.

## Best Practices

1. **Always use context with timeouts**:

   ```go
   ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
   defer cancel()
   ```

2. **Implement retry logic** for critical operations
3. **Use health notifications** to monitor RPC performance
4. **Close rotators** when done to prevent resource leaks
5. **Handle both HTTP and WebSocket** endpoints for redundancy
6. **Use batch requests** when making multiple calls

## Contributing

1. Fork the repository
2. Create your feature branch
3. Add tests for new functionality
4. Ensure all tests pass
5. Submit a pull request

## License

This project is licensed under the MIT License - see the LICENSE file for details.
