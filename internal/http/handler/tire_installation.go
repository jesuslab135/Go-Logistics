package handler

import (
	"context"

	"fleet/internal/db/gen"
	"fleet/internal/http/dto"
	"fleet/internal/http/middleware"
	"fleet/internal/platform/paginate"
)

type TireInstallationStore struct{ q *gen.Queries }

func NewTireInstallationStore(q *gen.Queries) *TireInstallationStore {
	return &TireInstallationStore{q: q}
}

func (s *TireInstallationStore) List(ctx context.Context, parentID int64, p paginate.Params) ([]dto.TireInstallationResponse, int64, error) {
	company := middleware.CompanyFromContext(ctx)
	rows, err := s.q.ListTireInstallations(ctx, gen.ListTireInstallationsParams{ParentID: parentID, CompanyID: company, Lim: int32(p.Limit), Off: int32(p.Offset)})
	if err != nil {
		return nil, 0, err
	}
	total, err := s.q.CountTireInstallations(ctx, gen.CountTireInstallationsParams{ParentID: parentID, CompanyID: company})
	if err != nil {
		return nil, 0, err
	}
	out := make([]dto.TireInstallationResponse, len(rows))
	for i, r := range rows {
		out[i] = toTireInstallationResponse(r)
	}
	return out, total, nil
}

func (s *TireInstallationStore) Get(ctx context.Context, parentID, id int64) (dto.TireInstallationResponse, error) {
	r, err := s.q.GetTireInstallation(ctx, gen.GetTireInstallationParams{ID: id, ParentID: parentID, CompanyID: middleware.CompanyFromContext(ctx)})
	if err != nil {
		return dto.TireInstallationResponse{}, err
	}
	return toTireInstallationResponse(r), nil
}

func (s *TireInstallationStore) Create(ctx context.Context, parentID int64, in dto.CreateTireInstallationRequest) (dto.TireInstallationResponse, error) {
	r, err := s.q.CreateTireInstallation(ctx, gen.CreateTireInstallationParams{
		ParentID:                 parentID,
		CompanyID:                middleware.CompanyFromContext(ctx),
		VehicleID:                in.VehicleID,
		PositionCode:             in.PositionCode,
		InstallDate:              in.InstallDate,
		OdometerAtInstall:        in.OdometerAtInstall,
		TreadDepthAtInstall32nds: in.TreadDepthAtInstall32nds,
		PsiAtInstall:             in.PsiAtInstall,
		InstalledByID:            in.InstalledByID,
		Notes:                    in.Notes,
	})
	if err != nil {
		return dto.TireInstallationResponse{}, err
	}
	return toTireInstallationResponse(r), nil
}

func (s *TireInstallationStore) Update(ctx context.Context, parentID, id int64, in dto.UpdateTireInstallationRequest) (dto.TireInstallationResponse, error) {
	r, err := s.q.UpdateTireInstallation(ctx, gen.UpdateTireInstallationParams{
		ID:                       id,
		ParentID:                 parentID,
		CompanyID:                middleware.CompanyFromContext(ctx),
		VehicleID:                in.VehicleID,
		PositionCode:             in.PositionCode,
		InstallDate:              in.InstallDate,
		OdometerAtInstall:        in.OdometerAtInstall,
		TreadDepthAtInstall32nds: in.TreadDepthAtInstall32nds,
		PsiAtInstall:             in.PsiAtInstall,
		InstalledByID:            in.InstalledByID,
		Notes:                    in.Notes,
	})
	if err != nil {
		return dto.TireInstallationResponse{}, err
	}
	return toTireInstallationResponse(r), nil
}

func (s *TireInstallationStore) Delete(ctx context.Context, parentID, id int64) error {
	return s.q.DeleteTireInstallation(ctx, gen.DeleteTireInstallationParams{ID: id, ParentID: parentID, CompanyID: middleware.CompanyFromContext(ctx)})
}

func toTireInstallationResponse(r gen.TireInstallation) dto.TireInstallationResponse {
	return dto.TireInstallationResponse{
		ID:                       r.ID,
		VehicleID:                r.VehicleID,
		TireID:                   r.TireID,
		PositionCode:             r.PositionCode,
		InstallDate:              r.InstallDate,
		OdometerAtInstall:        r.OdometerAtInstall,
		TreadDepthAtInstall32nds: r.TreadDepthAtInstall32nds,
		PsiAtInstall:             r.PsiAtInstall,
		InstalledByID:            r.InstalledByID,
		Notes:                    r.Notes,
	}
}
