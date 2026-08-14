package handler

import (
	"net/http"
	"sort"

	"github.com/gin-gonic/gin"

	"fleet/internal/db/gen"
	"fleet/internal/http/dto"
	"fleet/internal/http/middleware"
	"fleet/internal/platform/apierr"
)

// MeHandler serves the caller's own profile and authorization state. Django had
// no equivalent: the frontend inferred what a user could do by probing endpoints
// and reading 403s, which meant navigation could only be built after the fact.
type MeHandler struct{ q *gen.Queries }

func NewMeHandler(q *gen.Queries) *MeHandler { return &MeHandler{q: q} }

// Permissions godoc
//
//	@Summary		Current employee, effective permissions and companies
//	@Tags			me
//	@Security		BearerAuth
//	@Produce		json
//	@Success		200	{object}	dto.MePermissionsResponse
//	@Failure		401	{object}	dto.ErrorResponse
//	@Failure		403	{object}	dto.ErrorResponse
//	@Router			/api/v1/me/permissions [get]
func (h *MeHandler) Permissions(c *gin.Context) {
	identity, ok := middleware.IdentityOf(c)
	if !ok {
		apierr.Abort(c, apierr.Unauthorized("missing bearer token"))
		return
	}

	ctx := c.Request.Context()
	profile, err := h.q.GetMeProfile(ctx, identity.EmployeeID)
	if err != nil {
		apierr.Abort(c, err)
		return
	}
	companies, err := h.q.ListMyCompanies(ctx, identity.EmployeeID)
	if err != nil {
		apierr.Abort(c, err)
		return
	}

	out := dto.MePermissionsResponse{
		Employee: dto.MeEmployee{
			ID:                profile.ID,
			FirstName:         profile.FirstName,
			LastName:          profile.LastName,
			Email:             profile.Email,
			JobTitle:          profile.JobTitle,
			IsActive:          profile.IsActive,
			IsTechnician:      profile.IsTechnician,
			IsVehicleOperator: profile.IsVehicleOperator,
			DefaultCompanyID:  profile.DefaultCompanyID,
		},
		CompanyID:      identity.CompanyID,
		IsAdmin:        identity.IsAdmin,
		IsAccountOwner: identity.IsAccountOwner,
		Permissions:    identity.EffectivePermissions(),
		Modules:        readableModules(identity),
		Companies:      make([]dto.MeCompany, len(companies)),
	}

	if profile.RoleID != nil {
		name := ""
		if profile.RoleName != nil {
			name = *profile.RoleName
		}
		out.Role = &dto.MeRole{ID: *profile.RoleID, Name: name, IsAdmin: profile.RoleIsAdmin}
	}

	for i, company := range companies {
		out.Companies[i] = dto.MeCompany{ID: company.ID, Name: company.Name, Logo: company.Logo}
	}

	c.JSON(http.StatusOK, out)
}

// readableModules is the navigation filter: the modules the caller may read, in
// a stable order so the response does not churn between requests.
func readableModules(identity middleware.Identity) []string {
	modules := make([]string, 0, len(middleware.Modules))
	for _, module := range middleware.Modules {
		if identity.Can(module, http.MethodGet) {
			modules = append(modules, module)
		}
	}
	sort.Strings(modules)
	return modules
}
