package prices

import (
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/assets"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/common"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

const defaultPriceSource = "manual"

type assetReader interface {
	GetByID(assetID uuid.UUID) (*assets.Asset, error)
}

type priceWriterReader interface {
	Create(price *AssetPrice) error
	ListByAsset(assetID uuid.UUID, pagination common.Pagination) ([]AssetPrice, error)
	GetLatestByAsset(assetID uuid.UUID) (*AssetPrice, error)
}

type Service struct {
	repo      priceWriterReader
	assetRepo assetReader
}

func NewService(repo priceWriterReader, assetRepo assetReader) *Service {
	return &Service{
		repo:      repo,
		assetRepo: assetRepo,
	}
}

func (s *Service) Create(userID uuid.UUID, assetID uuid.UUID, req AssetPriceCreateRequest) (AssetPriceResponse, error) {
	asset, err := s.getAsset(assetID)
	if err != nil {
		return AssetPriceResponse{}, err
	}
	if !asset.IsActive {
		return AssetPriceResponse{}, common.BadRequest("Asset is inactive")
	}

	price, err := buildAssetPrice(userID, assetID, req)
	if err != nil {
		return AssetPriceResponse{}, err
	}

	if err := s.repo.Create(price); err != nil {
		return AssetPriceResponse{}, err
	}

	return ToResponse(*price), nil
}

func (s *Service) List(assetID uuid.UUID, pagination common.Pagination) ([]AssetPriceResponse, error) {
	if _, err := s.getAsset(assetID); err != nil {
		return nil, err
	}

	prices, err := s.repo.ListByAsset(assetID, pagination)
	if err != nil {
		return nil, err
	}

	responses := make([]AssetPriceResponse, 0, len(prices))
	for _, price := range prices {
		responses = append(responses, ToResponse(price))
	}
	return responses, nil
}

func (s *Service) GetLatest(assetID uuid.UUID) (AssetPriceResponse, error) {
	if _, err := s.getAsset(assetID); err != nil {
		return AssetPriceResponse{}, err
	}

	price, err := s.repo.GetLatestByAsset(assetID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return AssetPriceResponse{}, common.NotFound("Asset price not found")
	}
	if err != nil {
		return AssetPriceResponse{}, err
	}

	return ToResponse(*price), nil
}

func (s *Service) getAsset(assetID uuid.UUID) (*assets.Asset, error) {
	asset, err := s.assetRepo.GetByID(assetID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, common.NotFound("Asset not found")
	}
	if err != nil {
		return nil, err
	}
	if asset.AssetClass == assets.AssetClassFixedDeposit {
		return nil, common.NotFound("Asset not found")
	}
	return asset, nil
}

func buildAssetPrice(userID uuid.UUID, assetID uuid.UUID, req AssetPriceCreateRequest) (*AssetPrice, error) {
	if req.Price == nil || !req.Price.GreaterThan(decimal.Zero) {
		return nil, common.BadRequest("Price must be greater than zero")
	}

	currency := common.NormalizeCurrency(req.Currency)
	if !common.ValidateCurrency(currency) {
		return nil, common.BadRequest("Currency must be a three-letter uppercase code")
	}

	if req.PricedAt.IsZero() {
		return nil, common.BadRequest("Priced at timestamp is required")
	}
	if !common.ValidateNotFuture(req.PricedAt) {
		return nil, common.BadRequest("Priced at timestamp cannot be in the future")
	}

	source := strings.TrimSpace(req.Source)
	if source == "" {
		source = defaultPriceSource
	}

	return &AssetPrice{
		AssetID:         assetID,
		Price:           *req.Price,
		Currency:        currency,
		PricedAt:        req.PricedAt.UTC(),
		Source:          source,
		Note:            strings.TrimSpace(req.Note),
		CreatedByUserID: &userID,
	}, nil
}
