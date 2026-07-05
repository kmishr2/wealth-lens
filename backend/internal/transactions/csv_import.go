package transactions

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/common"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

const maxCSVImportRows = 1000

var csvImportHeader = []string{
	"transaction_type", "occurred_at", "description", "asset_id",
	"quantity", "amount", "currency", "idempotency_key",
}

type CSVImportResponse struct {
	RowsImported int `json:"rows_imported"`
}

func (s *Service) ImportCSV(userID uuid.UUID, portfolioID uuid.UUID, accountID uuid.UUID, reader io.Reader) (CSVImportResponse, error) {
	if _, err := s.getOwnedPortfolio(userID, portfolioID); err != nil {
		return CSVImportResponse{}, err
	}
	rows, err := parseCSVImport(reader, accountID)
	if err != nil {
		return CSVImportResponse{}, err
	}
	transactions := make([]*Transaction, 0, len(rows))
	for index, req := range rows {
		if _, err := s.repo.GetByIdempotencyKey(portfolioID, req.IdempotencyKey); err == nil {
			return CSVImportResponse{}, common.Conflict(fmt.Sprintf("Row %d idempotency key already exists", index+2))
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return CSVImportResponse{}, err
		}
		transaction, err := s.buildTransaction(userID, portfolioID, req, nil, nil, nil)
		if err != nil {
			return CSVImportResponse{}, common.BadRequest(fmt.Sprintf("Row %d: %s", index+2, err.Error()))
		}
		transactions = append(transactions, transaction)
	}
	if len(transactions) == 0 {
		return CSVImportResponse{}, common.BadRequest("CSV file contains no transaction rows")
	}
	if err := s.repo.CreateMany(transactions); err != nil {
		if common.IsUniqueViolation(err) {
			return CSVImportResponse{}, common.Conflict("CSV import contains an existing idempotency key")
		}
		return CSVImportResponse{}, err
	}
	return CSVImportResponse{RowsImported: len(transactions)}, nil
}

func parseCSVImport(input io.Reader, accountID uuid.UUID) ([]TransactionCreateRequest, error) {
	reader := csv.NewReader(input)
	reader.TrimLeadingSpace = true
	header, err := reader.Read()
	if err == io.EOF {
		return nil, common.BadRequest("CSV file is empty")
	}
	if err != nil {
		return nil, common.BadRequest("CSV header could not be read")
	}
	if len(header) != len(csvImportHeader) {
		return nil, common.BadRequest("CSV header does not match the required template")
	}
	for index := range header {
		header[index] = strings.TrimSpace(strings.TrimPrefix(header[index], "\ufeff"))
		if header[index] != csvImportHeader[index] {
			return nil, common.BadRequest("CSV header does not match the required template")
		}
	}

	requests := make([]TransactionCreateRequest, 0)
	keys := make(map[string]struct{})
	for rowNumber := 2; ; rowNumber++ {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, common.BadRequest(fmt.Sprintf("Row %d could not be parsed", rowNumber))
		}
		if len(requests) >= maxCSVImportRows {
			return nil, common.BadRequest(fmt.Sprintf("CSV import cannot exceed %d rows", maxCSVImportRows))
		}
		req, err := csvRowToRequest(row, accountID)
		if err != nil {
			return nil, common.BadRequest(fmt.Sprintf("Row %d: %s", rowNumber, err.Error()))
		}
		if _, exists := keys[req.IdempotencyKey]; exists {
			return nil, common.BadRequest(fmt.Sprintf("Row %d: duplicate idempotency key", rowNumber))
		}
		keys[req.IdempotencyKey] = struct{}{}
		requests = append(requests, req)
	}
	return requests, nil
}

func csvRowToRequest(row []string, accountID uuid.UUID) (TransactionCreateRequest, error) {
	transactionType := strings.ToLower(strings.TrimSpace(row[0]))
	if !common.OneOf(transactionType, TransactionTypeDeposit, TransactionTypeWithdrawal, TransactionTypeBuy, TransactionTypeSell, TransactionTypeFee, TransactionTypeTax) {
		return TransactionCreateRequest{}, errors.New("transaction_type must be deposit, withdrawal, buy, sell, fee, or tax")
	}
	occurredAt, err := time.Parse(time.RFC3339, strings.TrimSpace(row[1]))
	if err != nil {
		return TransactionCreateRequest{}, errors.New("occurred_at must use RFC3339 format")
	}
	currency := common.NormalizeCurrency(row[6])
	if !common.ValidateCurrency(currency) {
		return TransactionCreateRequest{}, errors.New("currency must be a three-letter code")
	}
	idempotencyKey := strings.TrimSpace(row[7])
	if idempotencyKey == "" {
		return TransactionCreateRequest{}, errors.New("idempotency_key is required")
	}
	amount, err := positiveCSVDecimal(row[5], "amount")
	if err != nil {
		return TransactionCreateRequest{}, err
	}
	entries := make([]TransactionEntryRequest, 0, 2)
	signedAmount := amount
	if common.OneOf(transactionType, TransactionTypeWithdrawal, TransactionTypeBuy, TransactionTypeFee, TransactionTypeTax) {
		signedAmount = amount.Neg()
	}
	if transactionType == TransactionTypeFee || transactionType == TransactionTypeTax {
		entries = append(entries, TransactionEntryRequest{EntryKind: transactionType, Amount: &signedAmount, Currency: currency})
	} else {
		entries = append(entries, TransactionEntryRequest{EntryKind: EntryKindCash, Amount: &signedAmount, Currency: currency})
	}
	if transactionType == TransactionTypeBuy || transactionType == TransactionTypeSell {
		assetID, err := uuid.Parse(strings.TrimSpace(row[3]))
		if err != nil {
			return TransactionCreateRequest{}, errors.New("asset_id is required for buy and sell rows")
		}
		quantity, err := positiveCSVDecimal(row[4], "quantity")
		if err != nil {
			return TransactionCreateRequest{}, err
		}
		if transactionType == TransactionTypeSell {
			quantity = quantity.Neg()
		}
		entries = append(entries, TransactionEntryRequest{EntryKind: EntryKindAsset, AssetID: &assetID, Quantity: &quantity, Currency: currency})
	} else if strings.TrimSpace(row[3]) != "" || strings.TrimSpace(row[4]) != "" {
		return TransactionCreateRequest{}, errors.New("asset_id and quantity are only allowed for buy and sell rows")
	}
	return TransactionCreateRequest{
		AccountID: accountID, TransactionType: transactionType, OccurredAt: occurredAt,
		Description: strings.TrimSpace(row[2]), IdempotencyKey: idempotencyKey, Entries: entries,
	}, nil
}

func positiveCSVDecimal(raw string, label string) (decimal.Decimal, error) {
	value, err := decimal.NewFromString(strings.TrimSpace(raw))
	if err != nil || !value.GreaterThan(decimal.Zero) {
		return decimal.Zero, fmt.Errorf("%s must be greater than zero", label)
	}
	return value, nil
}
