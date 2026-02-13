package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type CoinGecko struct {
	BaseURL string
	Client  *http.Client
}

func NewCoinGecko() (*CoinGecko, error) {
	if err := godotenv.Load(); err != nil {
		return nil, fmt.Errorf("error loading .env file: %w", err)
	}
	url := os.Getenv("URL")
	if url == "" {
		return nil, fmt.Errorf("URL not set in environment")
	}

	return &CoinGecko{
		BaseURL: url,
		Client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}, nil
}

func (cg *CoinGecko) FetchPrice(coinID string) (float64, error) {
	url := fmt.Sprintf("%s/simple/price?ids=%s&vs_currencies=usd", cg.BaseURL, coinID)

	resp, err := cg.Client.Get(url)

	if err != nil {
		return 0, fmt.Errorf("failed to fetch price for %s: %w", coinID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("API returned status %d for coin %s", resp.StatusCode, coinID)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("failed to read response body: %w", err)
	}

	var result map[string]map[string]float64

	if err := json.Unmarshal(body, &result); err != nil {
		return 0, fmt.Errorf("failed to parse JSON response: %w", err)
	}

	price, ok := result[coinID]["usd"]
	if !ok {
		return 0, fmt.Errorf("price not found for coin: %s", coinID)
	}

	return price, nil
}

func (cg *CoinGecko) FetchMultiplePrices(coinIDs ...string) (map[string]float64, error) {
	if len(coinIDs) == 0 {
		return nil, fmt.Errorf("no coin IDs provided")
	}

	coinList := strings.Join(coinIDs, ",")
	url := fmt.Sprintf("%s/simple/price?ids=%s&vs_currencies=usd", cg.BaseURL, coinList)

	resp, err := cg.Client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch prices: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var result map[string]map[string]float64
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	prices := make(map[string]float64)
	for _, coinID := range coinIDs {
		if data, ok := result[coinID]; ok {
			prices[coinID] = data["usd"]
		}
	}

	return prices, nil
}

func (cg *CoinGecko) GetSupportedCoins() (map[string]string, error) {
	url := fmt.Sprintf("%s/coins/markets?vs_currency=usd&order=market_cap_desc&per_page=100&page=1", cg.BaseURL)

	resp, err := cg.Client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch supported coins: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var coinList []struct {
		ID     string `json:"id"`
		Symbol string `json:"symbol"`
		Name   string `json:"name"`
	}

	if err := json.Unmarshal(body, &coinList); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	coins := make(map[string]string)
	for _, coin := range coinList {
		coins[coin.ID] = coin.Name
	}

	return coins, nil
}
