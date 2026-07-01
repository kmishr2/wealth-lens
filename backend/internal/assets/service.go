package assets

import (
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/common"
	"gorm.io/gorm"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(req AssetCreateRequest) (AssetResponse, error) {
	symbol := common.NormalizeSymbol(req.Symbol)
	if symbol == "" {
		return AssetResponse{}, common.BadRequest("Asset symbol is required")
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		return AssetResponse{}, common.BadRequest("Asset name is required")
	}

	assetClass := strings.ToLower(strings.TrimSpace(req.AssetClass))
	if !common.OneOf(assetClass, AssetClassCash, AssetClassEquity, AssetClassFund, AssetClassBond, AssetClassCrypto, AssetClassRealEstate, AssetClassCommodity, AssetClassAlternative, AssetClassOther) {
		return AssetResponse{}, common.BadRequest("Invalid asset class")
	}
	riskCategory, err := normalizeRiskCategory(req.RiskCategory, assetClass)
	if err != nil {
		return AssetResponse{}, err
	}

	currency := common.NormalizeCurrency(req.Currency)
	if !common.ValidateCurrency(currency) {
		return AssetResponse{}, common.BadRequest("Currency must be a three-letter uppercase code")
	}

	asset := &Asset{
		Symbol:       symbol,
		Name:         name,
		AssetClass:   assetClass,
		RiskCategory: riskCategory,
		Currency:     currency,
		Exchange:     strings.ToUpper(strings.TrimSpace(req.Exchange)),
		IsActive:     true,
	}

	if err := s.repo.Create(asset); err != nil {
		if common.IsUniqueViolation(err) {
			return AssetResponse{}, common.Conflict("Asset already exists")
		}
		return AssetResponse{}, err
	}

	return ToResponse(*asset), nil
}

func normalizeRiskCategory(raw *string, assetClass string) (*string, error) {
	if raw != nil && strings.TrimSpace(*raw) != "" {
		value := strings.ToLower(strings.TrimSpace(*raw))
		if !common.OneOf(value, RiskCategoryEquity, RiskCategoryDebt, RiskCategoryCashOther) {
			return nil, common.BadRequest("Risk category must be equity, debt, or cash_other")
		}
		return &value, nil
	}
	defaults := map[string]string{
		AssetClassEquity: RiskCategoryEquity,
		AssetClassBond:   RiskCategoryDebt,
		AssetClassCash:   RiskCategoryCashOther,
	}
	if value, ok := defaults[assetClass]; ok {
		return &value, nil
	}
	return nil, nil
}

func (s *Service) List(pagination common.Pagination) ([]AssetResponse, error) {
	assets, err := s.repo.List(pagination)
	if err != nil {
		return nil, err
	}

	responses := make([]AssetResponse, 0, len(assets))
	for _, asset := range assets {
		responses = append(responses, ToResponse(asset))
	}
	return responses, nil
}

func (s *Service) Get(assetID uuid.UUID) (AssetResponse, error) {
	asset, err := s.repo.GetByID(assetID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return AssetResponse{}, common.NotFound("Asset not found")
	}
	if err != nil {
		return AssetResponse{}, err
	}
	return ToResponse(*asset), nil
}
