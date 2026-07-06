package fixeddeposits

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/accounts"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/assets"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/common"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/portfolios"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/prices"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/transactions"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type bundleWriterReader interface {
	CreateBundle(asset *assets.Asset, transaction *transactions.Transaction, price *prices.AssetPrice, record *FixedDeposit) error
	ListByAccount(portfolioID uuid.UUID, accountID uuid.UUID) ([]FixedDeposit, error)
	GetByIDAccount(portfolioID uuid.UUID, accountID uuid.UUID, fixedDepositID uuid.UUID) (*FixedDeposit, error)
	CreateValue(price *prices.AssetPrice) error
	GetClosureByFixedDeposit(fixedDepositID uuid.UUID) (*Closure, error)
	CloseBundle(transaction *transactions.Transaction, closure *Closure) error
}

func (s *Service) CreateValue(userID uuid.UUID, portfolioID uuid.UUID, accountID uuid.UUID, fixedDepositID uuid.UUID, req ValueCreateRequest) (Response, error) {
	if _, err := s.getOwnedPortfolio(userID, portfolioID); err != nil {
		return Response{}, err
	}
	if _, err := s.getAccount(portfolioID, accountID); err != nil {
		return Response{}, err
	}
	record, err := s.repo.GetByIDAccount(portfolioID, accountID, fixedDepositID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return Response{}, common.NotFound("Fixed deposit not found")
	}
	if err != nil {
		return Response{}, err
	}
	if _, err := s.repo.GetClosureByFixedDeposit(record.ID); err == nil {
		return Response{}, common.Conflict("Closed fixed deposit values cannot be updated")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return Response{}, err
	}
	if req.CurrentValue == nil || !req.CurrentValue.GreaterThan(decimal.Zero) {
		return Response{}, common.BadRequest("Current value must be greater than zero")
	}
	valueDate, err := parseDate(req.CurrentValueDate, "Current value date")
	if err != nil {
		return Response{}, err
	}
	if valueDate.Before(record.StartDate) || valueDate.After(utcDate(time.Now().UTC())) {
		return Response{}, common.BadRequest("Current value date must be between start date and today")
	}
	price := &prices.AssetPrice{
		ID: uuid.New(), AssetID: record.AssetID, Price: req.CurrentValue.Round(10), Currency: record.Currency,
		PricedAt: valueDate, Source: "fixed-deposit-manual", Note: "Explicit fixed-deposit current value",
		CreatedByUserID: &userID,
	}
	if err := s.repo.CreateValue(price); err != nil {
		return Response{}, err
	}
	response := toResponse(*record, price.Price, price.PricedAt)
	applyLifecycle(&response, *record, nil, time.Now().UTC())
	return response, nil
}

type portfolioReader interface {
	GetOwned(userID uuid.UUID, portfolioID uuid.UUID) (*portfolios.Portfolio, error)
}

type accountReader interface {
	GetInPortfolio(portfolioID uuid.UUID, accountID uuid.UUID) (*accounts.Account, error)
}

type latestPriceReader interface {
	GetLatestByAsset(assetID uuid.UUID) (*prices.AssetPrice, error)
}

type Service struct {
	repo          bundleWriterReader
	portfolioRepo portfolioReader
	accountRepo   accountReader
	priceRepo     latestPriceReader
}

func NewService(repo bundleWriterReader, portfolioRepo portfolioReader, accountRepo accountReader, priceRepo latestPriceReader) *Service {
	return &Service{repo: repo, portfolioRepo: portfolioRepo, accountRepo: accountRepo, priceRepo: priceRepo}
}

func (s *Service) Create(userID uuid.UUID, portfolioID uuid.UUID, accountID uuid.UUID, req CreateRequest) (Response, error) {
	if _, err := s.getOwnedPortfolio(userID, portfolioID); err != nil {
		return Response{}, err
	}
	account, err := s.getAccount(portfolioID, accountID)
	if err != nil {
		return Response{}, err
	}
	if account.AccountType != accounts.AccountTypeBank {
		return Response{}, common.BadRequest("Fixed deposits can only be added to bank accounts")
	}

	validated, err := validateCreateRequest(req, account.Currency, time.Now().UTC())
	if err != nil {
		return Response{}, err
	}
	asset, transaction, price, record := buildBundle(userID, portfolioID, *account, validated)
	if err := s.repo.CreateBundle(asset, transaction, price, record); err != nil {
		if common.IsUniqueViolation(err) {
			return Response{}, common.Conflict("Fixed deposit already exists")
		}
		return Response{}, err
	}
	response := toResponse(*record, price.Price, price.PricedAt)
	applyLifecycle(&response, *record, nil, time.Now().UTC())
	return response, nil
}

func (s *Service) List(userID uuid.UUID, portfolioID uuid.UUID, accountID uuid.UUID) ([]Response, error) {
	if _, err := s.getOwnedPortfolio(userID, portfolioID); err != nil {
		return nil, err
	}
	if _, err := s.getAccount(portfolioID, accountID); err != nil {
		return nil, err
	}
	records, err := s.repo.ListByAccount(portfolioID, accountID)
	if err != nil {
		return nil, err
	}
	responses := make([]Response, 0, len(records))
	for _, record := range records {
		price, err := s.priceRepo.GetLatestByAsset(record.AssetID)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.Internal("Fixed deposit current value is missing")
		}
		if err != nil {
			return nil, err
		}
		var closure *Closure
		closure, err = s.repo.GetClosureByFixedDeposit(record.ID)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			closure = nil
		} else if err != nil {
			return nil, err
		}
		response := toResponse(record, price.Price, price.PricedAt)
		applyLifecycle(&response, record, closure, time.Now().UTC())
		responses = append(responses, response)
	}
	return responses, nil
}

func (s *Service) Close(userID uuid.UUID, portfolioID uuid.UUID, accountID uuid.UUID, fixedDepositID uuid.UUID, req CloseRequest) (Response, error) {
	if _, err := s.getOwnedPortfolio(userID, portfolioID); err != nil {
		return Response{}, err
	}
	if _, err := s.getAccount(portfolioID, accountID); err != nil {
		return Response{}, err
	}
	record, err := s.repo.GetByIDAccount(portfolioID, accountID, fixedDepositID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return Response{}, common.NotFound("Fixed deposit not found")
	}
	if err != nil {
		return Response{}, err
	}
	if _, err := s.repo.GetClosureByFixedDeposit(record.ID); err == nil {
		return Response{}, common.Conflict("Fixed deposit is already closed")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return Response{}, err
	}
	if req.Proceeds == nil || !req.Proceeds.GreaterThan(decimal.Zero) {
		return Response{}, common.BadRequest("Closing proceeds must be greater than zero")
	}
	closedAt, err := parseDate(req.ClosedAt, "Closed date")
	if err != nil {
		return Response{}, err
	}
	if closedAt.Before(record.StartDate) || closedAt.After(utcDate(time.Now().UTC())) {
		return Response{}, common.BadRequest("Closed date must be between start date and today")
	}
	closureType := strings.ToLower(strings.TrimSpace(req.ClosureType))
	expectedType := "maturity"
	if closedAt.Before(record.MaturityDate) {
		expectedType = "premature"
	}
	if closureType != expectedType {
		return Response{}, common.BadRequest("Closure type must match whether the closed date is before maturity")
	}
	transactionID := uuid.New()
	proceeds := req.Proceeds.Round(4)
	negativeOne := decimal.NewFromInt(-1)
	idempotencyKey := "fixed-deposit-close:" + record.ID.String()
	transaction := &transactions.Transaction{
		ID: transactionID, PortfolioID: portfolioID, AccountID: accountID,
		TransactionType: transactions.TransactionTypeSell, OccurredAt: closedAt,
		Description: "Closed fixed deposit: " + record.Name, IdempotencyKey: &idempotencyKey,
		CreatedByUserID: userID,
		Entries: []transactions.TransactionEntry{
			{ID: uuid.New(), TransactionID: transactionID, EntryKind: transactions.EntryKindAsset, AssetID: &record.AssetID, Quantity: &negativeOne, Currency: record.Currency},
			{ID: uuid.New(), TransactionID: transactionID, EntryKind: transactions.EntryKindCash, Amount: &proceeds, Currency: record.Currency},
		},
	}
	closure := &Closure{
		ID: uuid.New(), FixedDepositID: record.ID, PortfolioID: portfolioID, AccountID: accountID,
		ClosingTransactionID: transactionID, ClosureType: closureType, ClosedAt: closedAt,
		Proceeds: proceeds, Currency: record.Currency, Note: strings.TrimSpace(req.Note), CreatedByUserID: userID,
	}
	if err := s.repo.CloseBundle(transaction, closure); err != nil {
		if common.IsUniqueViolation(err) {
			return Response{}, common.Conflict("Fixed deposit is already closed")
		}
		return Response{}, err
	}
	response := toResponse(*record, closure.Proceeds, closure.ClosedAt)
	applyLifecycle(&response, *record, closure, time.Now().UTC())
	return response, nil
}

type validatedCreate struct {
	name               string
	bankReference      string
	principal          decimal.Decimal
	currency           string
	annualInterestRate decimal.Decimal
	startDate          time.Time
	maturityDate       time.Time
	currentValue       decimal.Decimal
	currentValueDate   time.Time
}

func validateCreateRequest(req CreateRequest, accountCurrency string, now time.Time) (validatedCreate, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return validatedCreate{}, common.BadRequest("Fixed deposit name is required")
	}
	if req.Principal == nil || !req.Principal.GreaterThan(decimal.Zero) {
		return validatedCreate{}, common.BadRequest("Principal must be greater than zero")
	}
	if req.CurrentValue == nil || !req.CurrentValue.GreaterThan(decimal.Zero) {
		return validatedCreate{}, common.BadRequest("Current value must be greater than zero")
	}
	if req.AnnualInterestRate == nil || !req.AnnualInterestRate.GreaterThan(decimal.Zero) || req.AnnualInterestRate.GreaterThan(decimal.NewFromInt(100)) {
		return validatedCreate{}, common.BadRequest("Annual interest rate must be greater than zero and no more than 100")
	}
	currency := common.NormalizeCurrency(req.Currency)
	if !common.ValidateCurrency(currency) {
		return validatedCreate{}, common.BadRequest("Currency must be a three-letter uppercase code")
	}
	if currency != accountCurrency {
		return validatedCreate{}, common.BadRequest("Fixed deposit currency must match account currency")
	}
	startDate, err := parseDate(req.StartDate, "Start date")
	if err != nil {
		return validatedCreate{}, err
	}
	maturityDate, err := parseDate(req.MaturityDate, "Maturity date")
	if err != nil {
		return validatedCreate{}, err
	}
	currentValueDate, err := parseDate(req.CurrentValueDate, "Current value date")
	if err != nil {
		return validatedCreate{}, err
	}
	today := utcDate(now)
	if startDate.After(today) {
		return validatedCreate{}, common.BadRequest("Start date cannot be in the future")
	}
	if !maturityDate.After(startDate) {
		return validatedCreate{}, common.BadRequest("Maturity date must be after start date")
	}
	if currentValueDate.Before(startDate) || currentValueDate.After(today) {
		return validatedCreate{}, common.BadRequest("Current value date must be between start date and today")
	}
	return validatedCreate{
		name:               name,
		bankReference:      strings.TrimSpace(req.BankReference),
		principal:          req.Principal.Round(4),
		currency:           currency,
		annualInterestRate: req.AnnualInterestRate.Round(6),
		startDate:          startDate,
		maturityDate:       maturityDate,
		currentValue:       req.CurrentValue.Round(10),
		currentValueDate:   currentValueDate,
	}, nil
}

func buildBundle(userID uuid.UUID, portfolioID uuid.UUID, account accounts.Account, input validatedCreate) (*assets.Asset, *transactions.Transaction, *prices.AssetPrice, *FixedDeposit) {
	assetID := uuid.New()
	transactionID := uuid.New()
	riskCategory := assets.RiskCategoryDebt
	asset := &assets.Asset{
		ID: assetID, Symbol: "FD-" + strings.ToUpper(strings.ReplaceAll(assetID.String(), "-", "")[:12]),
		Name: input.name, AssetClass: assets.AssetClassFixedDeposit, RiskCategory: &riskCategory,
		Currency: input.currency, Exchange: "BANK", IsActive: true,
	}
	negativePrincipal := input.principal.Neg()
	one := decimal.NewFromInt(1)
	transaction := &transactions.Transaction{
		ID: transactionID, PortfolioID: portfolioID, AccountID: account.ID,
		TransactionType: transactions.TransactionTypeBuy, OccurredAt: input.startDate,
		Description: "Opened fixed deposit: " + input.name, CreatedByUserID: userID,
		Entries: []transactions.TransactionEntry{
			{ID: uuid.New(), TransactionID: transactionID, EntryKind: transactions.EntryKindCash, Amount: &negativePrincipal, Currency: input.currency},
			{ID: uuid.New(), TransactionID: transactionID, EntryKind: transactions.EntryKindAsset, AssetID: &assetID, Quantity: &one, Currency: input.currency},
		},
	}
	price := &prices.AssetPrice{
		ID: uuid.New(), AssetID: assetID, Price: input.currentValue, Currency: input.currency,
		PricedAt: input.currentValueDate, Source: "fixed-deposit-manual",
		Note: "Explicit current value at fixed-deposit creation", CreatedByUserID: &userID,
	}
	record := &FixedDeposit{
		ID: uuid.New(), PortfolioID: portfolioID, AccountID: account.ID, AssetID: assetID,
		OpeningTransactionID: transactionID, Name: input.name, BankReference: input.bankReference,
		Principal: input.principal, Currency: input.currency, AnnualInterestRate: input.annualInterestRate,
		StartDate: input.startDate, MaturityDate: input.maturityDate, CreatedByUserID: userID,
	}
	return asset, transaction, price, record
}

func parseDate(raw string, label string) (time.Time, error) {
	date, err := time.Parse(dateLayout, strings.TrimSpace(raw))
	if err != nil {
		return time.Time{}, common.BadRequest(label + " must use YYYY-MM-DD format")
	}
	return utcDate(date), nil
}

func utcDate(date time.Time) time.Time {
	return time.Date(date.UTC().Year(), date.UTC().Month(), date.UTC().Day(), 0, 0, 0, 0, time.UTC)
}

func (s *Service) getOwnedPortfolio(userID uuid.UUID, portfolioID uuid.UUID) (*portfolios.Portfolio, error) {
	portfolio, err := s.portfolioRepo.GetOwned(userID, portfolioID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, common.NotFound("Portfolio not found")
	}
	return portfolio, err
}

func (s *Service) getAccount(portfolioID uuid.UUID, accountID uuid.UUID) (*accounts.Account, error) {
	account, err := s.accountRepo.GetInPortfolio(portfolioID, accountID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, common.NotFound("Account not found")
	}
	return account, err
}
