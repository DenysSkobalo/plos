package finance

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const (
	ISOEUR = 978
	ISOUAH = 980

	SourceMonobank = "MONO"
	SourcePUMB     = "PUMB"
	SourceManual   = "MANUAL"
)

type MonobankRateResponse struct {
	CurrencyCodeA int     `json:"currencyCodeA"`
	CurrencyCodeB int     `json:"currencyCodeB"`
	Date          int64   `json:"date"`
	RateBuy       float64 `json:"rateBuy"`
	RateSell      float64 `json:"rateSell"`
	RateCross     float64 `json:"rateCross"`
}

type RateService struct {
	repo       Repository
	httpClient *http.Client
	monoAPIURL string
}

func NewRateService(repo Repository, httpClient *http.Client) *RateService {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 10 * time.Second}
	}
	return &RateService{
		repo:       repo,
		httpClient: httpClient,
		monoAPIURL: "https://api.monobank.ua/bank/currency",
	}
}

func (s *RateService) SyncMonobankRates(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.monoAPIURL, http.NoBody)
	if err != nil {
		return fmt.Errorf("failed to create monobank request: %w", err)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("monobank api request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("monobank api returned status: %d", resp.StatusCode)
	}

	var rates []MonobankRateResponse
	if err := json.NewDecoder(resp.Body).Decode(&rates); err != nil {
		return fmt.Errorf("failed to decode monobank response: %w", err)
	}

	for _, r := range rates {
		if r.CurrencyCodeA == ISOEUR && r.CurrencyCodeB == ISOUAH {
			rate := ExchangeRate{
				Source:         SourceMonobank,
				BaseCurrency:   "EUR",
				TargetCurrency: "UAH",
				BuyRate:        r.RateBuy,
				SellRate:       r.RateSell,
				IsManual:       false,
			}
			return s.repo.SaveExchangeRate(ctx, &rate)
		}
	}

	return fmt.Errorf("EUR/UAH rate pair not found in monobank response")
}

func (s *RateService) SetManualRate(ctx context.Context, base, target string, buy, sell float64) error {
	rate := ExchangeRate{
		Source:         SourceManual,
		BaseCurrency:   base,
		TargetCurrency: target,
		BuyRate:        buy,
		SellRate:       sell,
		IsManual:       true,
	}
	return s.repo.SaveExchangeRate(ctx, &rate)
}

func (s *RateService) GetEffectiveRate(ctx context.Context, base, target string) (*ExchangeRate, error) {
	rates, err := s.repo.GetLatestRates(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to query latest rates: %w", err)
	}

	var bestMatch *ExchangeRate
	for i := range rates {
		r := rates[i]
		if r.BaseCurrency == base && r.TargetCurrency == target {
			if r.IsManual {
				return &r, nil
			}
			if bestMatch == nil {
				bestMatch = &r
			}
		}
	}

	if bestMatch != nil {
		return bestMatch, nil
	}

	return nil, fmt.Errorf("no exchange rate found for pair %s/%s", base, target)
}
