package fixeddeposits

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/accounts"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/assets"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/common"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/portfolios"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/prices"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/transactions"
	"github.com/shopspring/decimal"
)

type fakeBundleRepo struct {
	asset       *assets.Asset
	transaction *transactions.Transaction
	price       *prices.AssetPrice
	record      *FixedDeposit
}

func (f *fakeBundleRepo) CreateBundle(asset *assets.Asset, transaction *transactions.Transaction, price *prices.AssetPrice, record *FixedDeposit) error {
	f.asset, f.transaction, f.price, f.record = asset, transaction, price, record
	return nil
}

func (f *fakeBundleRepo) ListByAccount(uuid.UUID, uuid.UUID) ([]FixedDeposit, error) {
	return nil, nil
}

func (f *fakeBundleRepo) GetByIDAccount(uuid.UUID, uuid.UUID, uuid.UUID) (*FixedDeposit, error) {
	return f.record, nil
}

func (f *fakeBundleRepo) CreateValue(price *prices.AssetPrice) error {
	f.price = price
	return nil
}

type fakePortfolioRepo struct{ portfolio *portfolios.Portfolio }

func (f fakePortfolioRepo) GetOwned(uuid.UUID, uuid.UUID) (*portfolios.Portfolio, error) {
	if f.portfolio == nil {
		return nil, errors.New("portfolio missing")
	}
	return f.portfolio, nil
}

type fakeAccountRepo struct{ account *accounts.Account }

func (f fakeAccountRepo) GetInPortfolio(uuid.UUID, uuid.UUID) (*accounts.Account, error) {
	if f.account == nil {
		return nil, errors.New("account missing")
	}
	return f.account, nil
}

type fakePriceRepo struct{}

func (fakePriceRepo) GetLatestByAsset(uuid.UUID) (*prices.AssetPrice, error) {
	return nil, errors.New("not used")
}

func TestCreateBuildsAtomicLedgerBackedFixedDeposit(t *testing.T) {
	userID, portfolioID, accountID := uuid.New(), uuid.New(), uuid.New()
	repo := &fakeBundleRepo{}
	service := NewService(
		repo,
		fakePortfolioRepo{portfolio: &portfolios.Portfolio{ID: portfolioID, UserID: userID}},
		fakeAccountRepo{account: &accounts.Account{ID: accountID, PortfolioID: portfolioID, AccountType: accounts.AccountTypeBank, Currency: "INR"}},
		fakePriceRepo{},
	)
	principal := decimal.RequireFromString("100000")
	rate := decimal.RequireFromString("7.25")
	currentValue := decimal.RequireFromString("103500.50")

	response, err := service.Create(userID, portfolioID, accountID, CreateRequest{
		Name: "One-year FD", BankReference: "FD-123", Principal: &principal,
		Currency: "INR", AnnualInterestRate: &rate, StartDate: "2025-01-01",
		MaturityDate: "2027-01-01", CurrentValue: &currentValue, CurrentValueDate: "2026-01-01",
	})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if repo.asset == nil || repo.transaction == nil || repo.price == nil || repo.record == nil {
		t.Fatal("CreateBundle did not receive the complete fixed-deposit bundle")
	}
	if repo.asset.AssetClass != assets.AssetClassFixedDeposit || repo.asset.RiskCategory == nil || *repo.asset.RiskCategory != assets.RiskCategoryDebt {
		t.Fatalf("asset classification = %s/%v", repo.asset.AssetClass, repo.asset.RiskCategory)
	}
	if repo.transaction.TransactionType != transactions.TransactionTypeBuy || len(repo.transaction.Entries) != 2 {
		t.Fatalf("opening transaction = %s with %d entries", repo.transaction.TransactionType, len(repo.transaction.Entries))
	}
	if got := *repo.transaction.Entries[0].Amount; !got.Equal(principal.Neg()) {
		t.Fatalf("cash entry = %s, want %s", got, principal.Neg())
	}
	if got := *repo.transaction.Entries[1].Quantity; !got.Equal(decimal.NewFromInt(1)) {
		t.Fatalf("asset quantity = %s, want 1", got)
	}
	if !repo.price.Price.Equal(currentValue) || repo.price.Source != "fixed-deposit-manual" {
		t.Fatalf("price = %s/%s", repo.price.Price, repo.price.Source)
	}
	if response.AssetID != repo.asset.ID || response.OpeningTransactionID != repo.transaction.ID {
		t.Fatal("response links do not match the atomic bundle")
	}
	if response.ValuationMetadata.Formula == "" || len(response.ValuationMetadata.Assumptions) == 0 {
		t.Fatal("valuation metadata is incomplete")
	}
}

func TestCreateRejectsNonBankAccount(t *testing.T) {
	portfolioID := uuid.New()
	service := NewService(
		&fakeBundleRepo{},
		fakePortfolioRepo{portfolio: &portfolios.Portfolio{ID: portfolioID}},
		fakeAccountRepo{account: &accounts.Account{ID: uuid.New(), PortfolioID: portfolioID, AccountType: accounts.AccountTypeBrokerage, Currency: "INR"}},
		fakePriceRepo{},
	)

	_, err := service.Create(uuid.New(), portfolioID, uuid.New(), CreateRequest{})
	assertAppErrorMessage(t, err, "Fixed deposits can only be added to bank accounts")
}

func TestCreateValueAppendsExplicitObservation(t *testing.T) {
	userID, portfolioID, accountID := uuid.New(), uuid.New(), uuid.New()
	repo := &fakeBundleRepo{record: &FixedDeposit{
		ID: uuid.New(), PortfolioID: portfolioID, AccountID: accountID, AssetID: uuid.New(),
		Currency: "INR", StartDate: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
	}}
	service := NewService(
		repo,
		fakePortfolioRepo{portfolio: &portfolios.Portfolio{ID: portfolioID, UserID: userID}},
		fakeAccountRepo{account: &accounts.Account{ID: accountID, PortfolioID: portfolioID, AccountType: accounts.AccountTypeBank, Currency: "INR"}},
		fakePriceRepo{},
	)
	value := decimal.RequireFromString("110250.75")
	response, err := service.CreateValue(userID, portfolioID, accountID, repo.record.ID, ValueCreateRequest{
		CurrentValue: &value, CurrentValueDate: "2026-01-31",
	})
	if err != nil {
		t.Fatalf("CreateValue returned error: %v", err)
	}
	if repo.price == nil || !repo.price.Price.Equal(value) || repo.price.AssetID != repo.record.AssetID {
		t.Fatalf("appended price = %+v", repo.price)
	}
	if !response.CurrentValue.Equal(value) || response.CurrentValueAt.Format(dateLayout) != "2026-01-31" {
		t.Fatalf("response value = %s at %s", response.CurrentValue, response.CurrentValueAt)
	}
}

func TestValidateCreateRequestRejectsInvalidTerms(t *testing.T) {
	positive := decimal.NewFromInt(1)
	valid := CreateRequest{
		Name: "FD", Principal: &positive, Currency: "INR", AnnualInterestRate: &positive,
		StartDate: "2025-01-01", MaturityDate: "2027-01-01", CurrentValue: &positive,
		CurrentValueDate: "2026-01-01",
	}
	tests := []struct {
		name    string
		mutate  func(*CreateRequest)
		message string
	}{
		{"maturity before start", func(req *CreateRequest) { req.MaturityDate = "2024-01-01" }, "Maturity date must be after start date"},
		{"future start", func(req *CreateRequest) { req.StartDate = "2026-07-06" }, "Start date cannot be in the future"},
		{"value before start", func(req *CreateRequest) { req.CurrentValueDate = "2024-12-31" }, "Current value date must be between start date and today"},
		{"currency mismatch", func(req *CreateRequest) { req.Currency = "USD" }, "Fixed deposit currency must match account currency"},
	}
	now := time.Date(2026, 7, 5, 12, 0, 0, 0, time.UTC)
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := valid
			test.mutate(&req)
			_, err := validateCreateRequest(req, "INR", now)
			assertAppErrorMessage(t, err, test.message)
		})
	}
}

func assertAppErrorMessage(t *testing.T, err error, expected string) {
	t.Helper()
	var appErr *common.AppError
	if !errors.As(err, &appErr) || appErr.Message != expected {
		t.Fatalf("error = %v, want app error %q", err, expected)
	}
}
