package jsonrpc

import (
	"context"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/goccy/go-json"

	"github.com/go-resty/resty/v2"
	"github.com/rs/zerolog/log"
)

// ChainData holds all the information for a specific chain
type ChainData struct {
	PublicRPCs    []string
	NativeSymbol  string
	NativeDecimal int64
	ChainName     string
	Icon          string
}

var (
	mu       = &sync.RWMutex{}
	chains   = make(map[int64]*ChainData)
	isLoaded = &atomic.Bool{}
)

const (
	chainlistFilePath = "/tmp/chainlist.json"
	updateInterval    = 30 * time.Minute
)

type ChainlistResponse []ChainInfo

type ChainInfo struct {
	ChainID int    `json:"chainId"`
	Name    string `json:"name"`
	Icon    string `json:"icon"`
	RPC     []RPC  `json:"rpc"`
	Native  Native `json:"nativeCurrency"`
}

type RPC struct {
	URL string `json:"url"`
}

type Native struct {
	Name     string `json:"name"`
	Symbol   string `json:"symbol"`
	Decimals int    `json:"decimals"`
}

func init() {
	// Step 1: Try to load from offline file first
	if err := loadFromFile(); err != nil {
		log.Debug().Err(err).Msg("No offline chainlist file found")
	}
}

func loadFromFile() error {
	data, err := os.ReadFile(chainlistFilePath)
	if err != nil {
		return err
	}

	var response ChainlistResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return err
	}

	processChainlistData(response)

	isLoaded.Store(true)

	log.Info().Int("chains", len(chains)).Msg("Loaded chainlist from offline file")
	return nil
}

func ensureDataLoaded(ctx context.Context) error {
	if isLoaded.Load() {
		return nil
	}

	// Step 2: Lazy load from chainlist.org
	if err := loadFromChainlist(ctx); err != nil {
		log.Fatal().Err(err).Msg("Failed to load chainlist data - no offline file and unable to fetch from chainlist.org")
		return err
	}

	// Step 3: Start periodic update goroutine
	go func() {
		ticker := time.NewTicker(updateInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if err := loadFromChainlist(context.Background()); err != nil {
					log.Error().Err(err).Msg("Failed to update chainlist data")
				}

			case <-ctx.Done():
				return
			}
		}
	}()

	return nil
}

func loadFromChainlist(ctx context.Context) error {
	log.Info().Msg("Fetching chainlist from chainlist.org")

	body, err := resty.New().R().SetContext(ctx).Get("https://chainlist.org/rpcs.json")
	if err != nil {
		log.Error().Err(err).Msg("Failed to fetch public RPCs")
		return err
	}

	var response ChainlistResponse
	if err := json.Unmarshal(body.Body(), &response); err != nil {
		log.Error().Err(err).Msg("Failed to unmarshal public RPCs response")
		return err
	}

	// Persist only the parsed response struct to offline file
	cleanedData, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		log.Warn().Err(err).Msg("Failed to marshal cleaned chainlist data")
	} else {
		if err := os.WriteFile(chainlistFilePath, cleanedData, 0o644); err != nil {
			log.Warn().Err(err).Msg("Failed to persist chainlist to offline file")
		} else {
			log.Info().Msg("Persisted cleaned chainlist to offline file")
		}
	}

	processChainlistData(response)

	isLoaded.Store(true)

	log.Info().Int("chains", len(chains)).Msg("Loaded and cached chainlist from API")
	return nil
}

func processChainlistData(response ChainlistResponse) {
	mu.Lock()
	defer mu.Unlock()

	// Clear existing data
	chains = make(map[int64]*ChainData)

	for _, item := range response {
		if item.ChainID <= 0 {
			continue
		}

		var urls []string
		for _, rpc := range item.RPC {
			// Avoid url with place holder
			if strings.Contains(rpc.URL, "$") {
				continue
			}

			if rpc.URL != "" {
				urls = append(urls, rpc.URL)
			}
		}

		chains[int64(item.ChainID)] = &ChainData{
			PublicRPCs:    urls,
			NativeSymbol:  strings.ToUpper(item.Native.Symbol),
			NativeDecimal: int64(item.Native.Decimals),
			ChainName:     item.Name,
			Icon:          item.Icon,
		}
	}
}

func GetChainData(ctx context.Context, chainID int64) (*ChainData, error) {
	if err := ensureDataLoaded(ctx); err != nil {
		return nil, err
	}

	mu.RLock()
	defer mu.RUnlock()

	if chainData, exists := chains[chainID]; exists {
		// Return a copy to prevent external modifications
		return &ChainData{
			PublicRPCs:    append([]string{}, chainData.PublicRPCs...),
			NativeSymbol:  chainData.NativeSymbol,
			NativeDecimal: chainData.NativeDecimal,
			ChainName:     chainData.ChainName,
			Icon:          chainData.Icon,
		}, nil
	}
	return nil, nil
}

func GetAllChains(ctx context.Context) (map[int64]*ChainData, error) {
	if err := ensureDataLoaded(ctx); err != nil {
		return nil, err
	}

	mu.RLock()
	defer mu.RUnlock()

	result := make(map[int64]*ChainData)
	for chainID, chainData := range chains {
		// Return copies to prevent external modifications
		result[chainID] = &ChainData{
			PublicRPCs:    append([]string{}, chainData.PublicRPCs...),
			NativeSymbol:  chainData.NativeSymbol,
			NativeDecimal: chainData.NativeDecimal,
			ChainName:     chainData.ChainName,
			Icon:          chainData.Icon,
		}
	}
	return result, nil
}
