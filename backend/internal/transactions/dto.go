package transactions

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type TransactionEntryRequest struct {
	EntryKind string           `json:"entry_kind"`
	AssetID   *uuid.UUID       `json:"asset_id"`
	Quantity  *decimal.Decimal `json:"quantity"`
	Amount    *decimal.Decimal `json:"amount"`
	Currency  string           `json:"currency"`
}

type TransactionCreateRequest struct {
	AccountID       uuid.UUID                 `json:"account_id"`
	TransactionType string                    `json:"transaction_type"`
	OccurredAt      time.Time                 `json:"occurred_at"`
	Description     string                    `json:"description"`
	IdempotencyKey  string                    `json:"idempotency_key"`
	Entries         []TransactionEntryRequest `json:"entries"`
}

type TransactionReversalRequest struct {
	Reason     string    `json:"reason"`
	OccurredAt time.Time `json:"occurred_at"`
}

type TransactionCorrectionRequest struct {
	Reason      string                   `json:"reason"`
	Replacement TransactionCreateRequest `json:"replacement"`
}

type TransactionEntryResponse struct {
	ID        uuid.UUID        `json:"id"`
	EntryKind string           `json:"entry_kind"`
	AssetID   *uuid.UUID       `json:"asset_id,omitempty"`
	Quantity  *decimal.Decimal `json:"quantity,omitempty"`
	Amount    *decimal.Decimal `json:"amount,omitempty"`
	Currency  string           `json:"currency"`
	CreatedAt time.Time        `json:"created_at"`
}

type TransactionResponse struct {
	ID                    uuid.UUID                  `json:"id"`
	PortfolioID           uuid.UUID                  `json:"portfolio_id"`
	AccountID             uuid.UUID                  `json:"account_id"`
	TransactionType       string                     `json:"transaction_type"`
	OccurredAt            time.Time                  `json:"occurred_at"`
	Description           string                     `json:"description"`
	Entries               []TransactionEntryResponse `json:"entries"`
	ReversesTransactionID *uuid.UUID                 `json:"reverses_transaction_id,omitempty"`
	CorrectsTransactionID *uuid.UUID                 `json:"corrects_transaction_id,omitempty"`
	CreatedAt             time.Time                  `json:"created_at"`
}

type TransactionCorrectionResponse struct {
	Reversal    TransactionResponse `json:"reversal"`
	Replacement TransactionResponse `json:"replacement"`
}

func ToResponse(transaction Transaction) TransactionResponse {
	entries := make([]TransactionEntryResponse, 0, len(transaction.Entries))
	for _, entry := range transaction.Entries {
		entries = append(entries, TransactionEntryResponse{
			ID:        entry.ID,
			EntryKind: entry.EntryKind,
			AssetID:   entry.AssetID,
			Quantity:  entry.Quantity,
			Amount:    entry.Amount,
			Currency:  entry.Currency,
			CreatedAt: entry.CreatedAt,
		})
	}

	return TransactionResponse{
		ID:                    transaction.ID,
		PortfolioID:           transaction.PortfolioID,
		AccountID:             transaction.AccountID,
		TransactionType:       transaction.TransactionType,
		OccurredAt:            transaction.OccurredAt,
		Description:           transaction.Description,
		Entries:               entries,
		ReversesTransactionID: transaction.ReversesTransactionID,
		CorrectsTransactionID: transaction.CorrectsTransactionID,
		CreatedAt:             transaction.CreatedAt,
	}
}
