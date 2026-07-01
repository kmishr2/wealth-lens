package benchmarks

import (
	"time"

	"github.com/google/uuid"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/common"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(benchmark *Benchmark) error {
	return r.db.Create(benchmark).Error
}

func (r *Repository) List(pagination common.Pagination) ([]Benchmark, error) {
	var benchmarks []Benchmark
	err := r.db.
		Order("code asc, created_at asc, id asc").
		Limit(pagination.Limit).
		Offset(pagination.Offset).
		Find(&benchmarks).Error
	return benchmarks, err
}

func (r *Repository) GetByID(benchmarkID uuid.UUID) (*Benchmark, error) {
	var benchmark Benchmark
	err := r.db.First(&benchmark, "id = ?", benchmarkID).Error
	return &benchmark, err
}

func (r *Repository) CreateObservation(observation *BenchmarkObservation) error {
	return r.db.Create(observation).Error
}

func (r *Repository) ListObservations(benchmarkID uuid.UUID, pagination common.Pagination) ([]BenchmarkObservation, error) {
	var observations []BenchmarkObservation
	err := r.db.
		Where("benchmark_id = ?", benchmarkID).
		Order("observation_date desc, created_at desc, id desc").
		Limit(pagination.Limit).
		Offset(pagination.Offset).
		Find(&observations).Error
	return observations, err
}

func (r *Repository) GetObservationByDate(benchmarkID uuid.UUID, observationDate time.Time) (*BenchmarkObservation, error) {
	var observation BenchmarkObservation
	err := r.db.
		Where("benchmark_id = ? AND observation_date = ?", benchmarkID, observationDate).
		First(&observation).Error
	return &observation, err
}

func (r *Repository) ListObservationsByDateRange(benchmarkID uuid.UUID, startDate, endDate time.Time) ([]BenchmarkObservation, error) {
	var observations []BenchmarkObservation
	err := r.db.Where("benchmark_id = ? AND observation_date >= ? AND observation_date <= ?", benchmarkID, startDate, endDate).
		Order("observation_date asc, created_at asc, id asc").Find(&observations).Error
	return observations, err
}
