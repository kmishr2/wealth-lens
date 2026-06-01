package prices

import (
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/assets"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/common"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type fakeAssetReader struct {
	assetID uuid.UUID
	asset   *assets.Asset
	err     error
}

func (f *fakeAssetReader) GetByID(assetID uuid.UUID) (*assets.Asset, error) {
	f.assetID = assetID
	return f.asset, f.err
}

type fakePriceRepository struct {
	created     *AssetPrice
	createErr   error
	listAssetID uuid.UUID
	listPrices  []AssetPrice
	listErr     error
	latestID    uuid.UUID
	latestPrice *AssetPrice
	latestErr   error
}

func (f *fakePriceRepository) Create(price *AssetPrice) error {
	f.created = price
	return f.createErr
}

func (f *fakePriceRepository) ListByAsset(assetID uuid.UUID, pagination common.Pagination) ([]AssetPrice, error) {
	f.listAssetID = assetID
	return f.listPrices, f.listErr
}

func (f *fakePriceRepository) GetLatestByAsset(assetID uuid.UUID) (*AssetPrice, error) {
	f.latestID = assetID
	return f.latestPrice, f.latestErr
}

func TestCreateAssetPriceValidatesAndDefaultsManualSource(t *testing.T) {
	userID := uuid.New()
	assetID := uuid.New()
	price := decimal.RequireFromString("123.4567")
	pricedAt := time.Now().UTC().Add(-time.Hour)
	repo := &fakePriceRepository{}
	service := NewService(repo, &fakeAssetReader{
		asset: &assets.Asset{ID: assetID, IsActive: true},
	})

	response, err := service.Create(userID, assetID, AssetPriceCreateRequest{
		Price:    &price,
		Currency: "usd",
		PricedAt: pricedAt,
		Note:     " close ",
	})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	if repo.created == nil {
		t.Fatal("repo.Create was not called")
	}
	if repo.created.AssetID != assetID || repo.created.CreatedByUserID != userID {
		t.Fatalf("created price audit fields = %+v", repo.created)
	}
	if repo.created.Source != defaultPriceSource {
		t.Fatalf("source = %q, want %q", repo.created.Source, defaultPriceSource)
	}
	if repo.created.Currency != "USD" {
		t.Fatalf("currency = %q, want USD", repo.created.Currency)
	}
	if repo.created.Note != "close" {
		t.Fatalf("note = %q, want close", repo.created.Note)
	}
	if !response.Price.Equal(price) {
		t.Fatalf("response price = %s, want %s", response.Price, price)
	}
}

func TestCreateAssetPriceRejectsInvalidInputs(t *testing.T) {
	assetID := uuid.New()
	validPrice := decimal.RequireFromString("1")
	future := time.Now().UTC().Add(time.Hour)

	tests := []struct {
		name        string
		req         AssetPriceCreateRequest
		wantMessage string
	}{
		{
			name: "missing price",
			req: AssetPriceCreateRequest{
				Currency: "USD",
				PricedAt: time.Now().UTC(),
			},
			wantMessage: "greater than zero",
		},
		{
			name: "bad currency",
			req: AssetPriceCreateRequest{
				Price:    &validPrice,
				Currency: "US1",
				PricedAt: time.Now().UTC(),
			},
			wantMessage: "three-letter uppercase code",
		},
		{
			name: "future priced at",
			req: AssetPriceCreateRequest{
				Price:    &validPrice,
				Currency: "USD",
				PricedAt: future,
			},
			wantMessage: "cannot be in the future",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewService(&fakePriceRepository{}, &fakeAssetReader{
				asset: &assets.Asset{ID: assetID, IsActive: true},
			})

			_, err := service.Create(uuid.New(), assetID, tt.req)
			if err == nil {
				t.Fatal("Create returned nil error")
			}
			if !strings.Contains(err.Error(), tt.wantMessage) {
				t.Fatalf("error = %q, want to contain %q", err.Error(), tt.wantMessage)
			}
		})
	}
}

func TestGetLatestReturnsNotFoundWhenNoPriceExists(t *testing.T) {
	service := NewService(
		&fakePriceRepository{latestErr: gorm.ErrRecordNotFound},
		&fakeAssetReader{asset: &assets.Asset{ID: uuid.New(), IsActive: true}},
	)

	_, err := service.GetLatest(uuid.New())
	if err == nil {
		t.Fatal("GetLatest returned nil error")
	}

	var appErr *common.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("error type = %T, want *common.AppError", err)
	}
	if appErr.Status != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", appErr.Status, http.StatusNotFound)
	}
}
