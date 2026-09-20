package http

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"plos/internal/domain/finance"
)

type Server struct {
	router        *chi.Mux
	repo          finance.Repository
	rateService   *finance.RateService
	engine        *finance.CashflowEngine
	exportService *finance.ExportService
}

func NewServer(
	repo finance.Repository,
	rateService *finance.RateService,
	engine *finance.CashflowEngine,
	exportService *finance.ExportService,
) *Server {
	s := &Server{
		router:        chi.NewRouter(),
		repo:          repo,
		rateService:   rateService,
		engine:        engine,
		exportService: exportService,
	}

	s.routes()
	return s
}

func (s *Server) Router() http.Handler {
	return s.router
}

func (s *Server) routes() {
	s.router.Use(middleware.RequestID)
	s.router.Use(middleware.Logger)
	s.router.Use(middleware.Recoverer)

	s.router.Route("/api/v1", func(r chi.Router) {
		r.Get("/currencies", s.handleGetCurrencies)
		r.Post("/currencies", s.handleCreateCurrency)

		r.Get("/rates", s.handleGetRates)
		r.Post("/rates/sync", s.handleSyncRates)
		r.Post("/rates/manual", s.handleSetManualRate)

		r.Get("/accounts", s.handleGetAccounts)
		r.Post("/accounts", s.handleCreateAccount)
		r.Get("/debts", s.handleGetDebts)
		r.Post("/debts", s.handleSetDebtMetadata)

		r.Post("/cashflow/simulate", s.handleSimulateCashflow)
		r.Get("/export/debts.csv", s.handleExportDebtsCSV)
	})
}

func (s *Server) handleGetCurrencies(w http.ResponseWriter, r *http.Request) {
	currencies, err := s.repo.GetCurrencies(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	respondJSON(w, http.StatusOK, currencies)
}

func (s *Server) handleCreateCurrency(w http.ResponseWriter, r *http.Request) {
	var c finance.Currency
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}

	if err := s.repo.AddCurrency(r.Context(), c); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}
func (s *Server) handleGetRates(w http.ResponseWriter, r *http.Request) {
	rates, err := s.repo.GetLatestRates(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	respondJSON(w, http.StatusOK, rates)
}

func (s *Server) handleSyncRates(w http.ResponseWriter, r *http.Request) {
	if err := s.rateService.SyncMonobankRates(r.Context()); err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	w.WriteHeader(http.StatusOK)
}

type setManualRateRequest struct {
	BaseCurrency   string  `json:"base_currency"`
	TargetCurrency string  `json:"target_currency"`
	BuyRate        float64 `json:"buy_rate"`
	SellRate       float64 `json:"sell_rate"`
}

func (s *Server) handleSetManualRate(w http.ResponseWriter, r *http.Request) {
	var req setManualRateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}

	err := s.rateService.SetManualRate(r.Context(), req.BaseCurrency, req.TargetCurrency, req.BuyRate, req.SellRate)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleGetAccounts(w http.ResponseWriter, r *http.Request) {
	accounts, err := s.repo.GetAccounts(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	respondJSON(w, http.StatusOK, accounts)
}

func (s *Server) handleCreateAccount(w http.ResponseWriter, r *http.Request) {
	var acc finance.Account
	if err := json.NewDecoder(r.Body).Decode(&acc); err != nil {
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}

	if err := s.repo.CreateAccount(r.Context(), &acc); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (s *Server) handleGetDebts(w http.ResponseWriter, r *http.Request) {
	debts, err := s.repo.GetActiveDebtsOrderedByPriority(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	respondJSON(w, http.StatusOK, debts)
}

func (s *Server) handleSetDebtMetadata(w http.ResponseWriter, r *http.Request) {
	var meta finance.DebtMetadata
	if err := json.NewDecoder(r.Body).Decode(&meta); err != nil {
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}

	if err := s.repo.SetDebtMetadata(r.Context(), &meta); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

type cashflowSimulateRequest struct {
	MonthName string                `json:"month_name"`
	Incomes   []finance.IncomeItem  `json:"incomes"`
	Expenses  []finance.ExpenseItem `json:"expenses"`
	BuyRate   float64               `json:"buy_rate"`
}

func (s *Server) handleSimulateCashflow(w http.ResponseWriter, r *http.Request) {
	var req cashflowSimulateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}

	activeDebts, err := s.repo.GetActiveDebtsOrderedByPriority(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	states := make([]finance.DebtPayoffState, len(activeDebts))
	for i, d := range activeDebts {
		acc, _ := s.repo.GetAccountByID(r.Context(), d.AccountID)
		balance := d.OriginalAmount
		if acc != nil && acc.CurrentBalance < 0 {
			balance = -acc.CurrentBalance
		}

		states[i] = finance.DebtPayoffState{
			AccountID:            d.AccountID,
			CreditorName:         d.CreditorName,
			Priority:             d.Priority,
			OriginalCurrency:     d.OriginalCurrency,
			RemainingBalance:     balance,
			MinMonthlyPaymentUAH: d.MinMonthlyPaymentUAH,
			Status:               d.Status,
		}
	}

	projection, _ := s.engine.SimulateMonth(req.MonthName, req.Incomes, req.Expenses, states, req.BuyRate)
	respondJSON(w, http.StatusOK, projection)
}

func (s *Server) handleExportDebtsCSV(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", `attachment; filename="debts.csv"`)

	if err := s.exportService.ExportDebtsToCSV(r.Context(), w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func respondJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}
