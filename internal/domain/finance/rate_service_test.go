package finance

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRateService_SyncAndResolutionPriority(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	_ = repo.AddCurrency(ctx, Currency{Code: "EUR", Symbol: "€"})
	_ = repo.AddCurrency(ctx, Currency{Code: "UAH", Symbol: "₴"})

	// Mock Monobank Server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[
			{"currencyCodeA": 978, "currencyCodeB": 980, "date": 1700000000, "rateBuy": 51.0700, "rateSell": 51.7706}
		]`))
	}))
	defer mockServer.Close()

	service := NewRateService(repo, mockServer.Client())
	service.monoAPIURL = mockServer.URL

	// 1. Синхронізація з Monobank API
	if err := service.SyncMonobankRates(ctx); err != nil {
		t.Fatalf("SyncMonobankRates failed: %v", err)
	}

	effectiveRate, err := service.GetEffectiveRate(ctx, "EUR", "UAH")
	if err != nil {
		t.Fatalf("GetEffectiveRate failed: %v", err)
	}
	if effectiveRate.Source != SourceMonobank || effectiveRate.BuyRate != 51.0700 {
		t.Errorf("expected Monobank rate 51.0700, got source %s with rate %f", effectiveRate.Source, effectiveRate.BuyRate)
	}

	// 2. Встановлення Manual Override
	if err := service.SetManualRate(ctx, "EUR", "UAH", 52.0000, 52.5000); err != nil {
		t.Fatalf("SetManualRate failed: %v", err)
	}

	effectiveRate, err = service.GetEffectiveRate(ctx, "EUR", "UAH")
	if err != nil {
		t.Fatalf("GetEffectiveRate failed after manual override: %v", err)
	}
	if !effectiveRate.IsManual || effectiveRate.BuyRate != 52.0000 {
		t.Errorf("expected Manual rate 52.0000 with IsManual=true, got source %s rate %f", effectiveRate.Source, effectiveRate.BuyRate)
	}
}
