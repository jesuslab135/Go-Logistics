package handler

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"fleet/internal/db/gen"
	"fleet/internal/http/dto"
	"fleet/internal/platform/paginate"
)

// CompanyStore adapts the sqlc-generated queries to crud.Store, translating
// between request/response DTOs and the persistence rows. Company creation
// seeds the 6 default roles transactionally.
type CompanyStore struct {
	q    *gen.Queries
	pool *pgxpool.Pool
}

func NewCompanyStore(q *gen.Queries, pool *pgxpool.Pool) *CompanyStore {
	return &CompanyStore{q: q, pool: pool}
}

func (s *CompanyStore) List(ctx context.Context, p paginate.Params) ([]dto.CompanyResponse, int64, error) {
	rows, err := s.q.ListCompanies(ctx, gen.ListCompaniesParams{
		Limit:  int32(p.Limit),
		Offset: int32(p.Offset),
	})
	if err != nil {
		return nil, 0, err
	}
	total, err := s.q.CountCompanies(ctx)
	if err != nil {
		return nil, 0, err
	}

	out := make([]dto.CompanyResponse, len(rows))
	for i, r := range rows {
		out[i] = toCompanyResponse(r)
	}
	return out, total, nil
}

func (s *CompanyStore) Get(ctx context.Context, id int64) (dto.CompanyResponse, error) {
	row, err := s.q.GetCompany(ctx, id)
	if err != nil {
		return dto.CompanyResponse{}, err
	}
	return toCompanyResponse(row), nil
}

func (s *CompanyStore) Create(ctx context.Context, in dto.CreateCompanyRequest) (dto.CompanyResponse, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return dto.CompanyResponse{}, err
	}
	defer tx.Rollback(ctx)

	qtx := s.q.WithTx(tx)

	row, err := qtx.CreateCompany(ctx, gen.CreateCompanyParams{
		Name:                in.Name,
		TaxID:               in.TaxID,
		Address:             in.Address,
		CreatedAt:           time.Now().UTC(),
		Phone:               in.Phone,
		Email:               in.Email,
		Website:             in.Website,
		Logo:                in.Logo,
		City:                in.City,
		Region:              in.Region,
		PostalCode:          in.PostalCode,
		Country:             orDefault(in.Country, "MX"),
		Timezone:            orDefault(in.Timezone, "America/Mexico_City"),
		Currency:            orDefault(in.Currency, "MXN"),
		SystemOfMeasurement: orDefault(in.SystemOfMeasurement, "metric"),
	})
	if err != nil {
		return dto.CompanyResponse{}, err
	}

	if err := SeedDefaultRoles(ctx, qtx, row.ID); err != nil {
		return dto.CompanyResponse{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return dto.CompanyResponse{}, err
	}

	return toCompanyResponse(row), nil
}

func (s *CompanyStore) Update(ctx context.Context, id int64, in dto.UpdateCompanyRequest) (dto.CompanyResponse, error) {
	row, err := s.q.UpdateCompany(ctx, gen.UpdateCompanyParams{
		ID:                  id,
		Name:                in.Name,
		TaxID:               in.TaxID,
		Address:             in.Address,
		Phone:               in.Phone,
		Email:               in.Email,
		Website:             in.Website,
		Logo:                in.Logo,
		City:                in.City,
		Region:              in.Region,
		PostalCode:          in.PostalCode,
		Country:             orDefault(in.Country, "MX"),
		Timezone:            orDefault(in.Timezone, "America/Mexico_City"),
		Currency:            orDefault(in.Currency, "MXN"),
		SystemOfMeasurement: orDefault(in.SystemOfMeasurement, "metric"),
	})
	if err != nil {
		return dto.CompanyResponse{}, err
	}
	return toCompanyResponse(row), nil
}

func (s *CompanyStore) Delete(ctx context.Context, id int64) error {
	return s.q.DeleteCompany(ctx, id)
}

func toCompanyResponse(c gen.Company) dto.CompanyResponse {
	return dto.CompanyResponse{
		ID:                  c.ID,
		Name:                c.Name,
		TaxID:               c.TaxID,
		Address:             c.Address,
		Phone:               c.Phone,
		Email:               c.Email,
		Website:             c.Website,
		Logo:                c.Logo,
		City:                c.City,
		Region:              c.Region,
		PostalCode:          c.PostalCode,
		Country:             c.Country,
		Timezone:            c.Timezone,
		Currency:            c.Currency,
		SystemOfMeasurement: c.SystemOfMeasurement,
		CreatedAt:           c.CreatedAt,
	}
}

func orDefault(v, fallback string) string {
	if v == "" {
		return fallback
	}
	return v
}
