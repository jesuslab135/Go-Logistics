package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"

	"fleet/internal/auth"
	"fleet/internal/config"
	"fleet/internal/db"
	"fleet/internal/db/gen"
	"fleet/internal/http/handler"
)

func main() {
	_ = godotenv.Load()

	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}

	var err error
	switch os.Args[1] {
	case "setpass":
		err = setpass(os.Args[2:])
	case "createuser":
		err = createuser(os.Args[2:])
	default:
		usage()
		os.Exit(2)
	}

	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

// setpass sets an employee's bcrypt password so they can log in. The employee
// must exist and be active.
func setpass(args []string) error {
	fs := flag.NewFlagSet("setpass", flag.ExitOnError)
	email := fs.String("email", "", "employee email")
	password := fs.String("password", "", "new password (min 8 chars)")
	_ = fs.Parse(args)

	if *email == "" || *password == "" {
		return fmt.Errorf("--email and --password are required")
	}
	if len(*password) < 8 {
		return fmt.Errorf("password must be at least 8 characters")
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	ctx := context.Background()
	pool, err := db.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()

	q := gen.New(pool)
	emp, err := q.GetEmployeeAuthByEmail(ctx, *email)
	if err != nil {
		return fmt.Errorf("no active employee for email %q: %w", *email, err)
	}

	hash, err := auth.HashPassword(*password)
	if err != nil {
		return err
	}
	if err := q.UpdateEmployeePassword(ctx, gen.UpdateEmployeePasswordParams{
		ID:           emp.ID,
		PasswordHash: hash,
		UpdatedAt:    time.Now().UTC(),
	}); err != nil {
		return err
	}

	fmt.Printf("password updated for employee %d (%s)\n", emp.ID, *email)
	return nil
}

// createuser creates a new employee with a password. It creates a company and
// admin role by default, or links to an existing company if --company-id is set.
func createuser(args []string) error {
	fs := flag.NewFlagSet("createuser", flag.ExitOnError)
	email := fs.String("email", "", "employee email (required)")
	password := fs.String("password", "", "password, min 8 chars (required)")
	firstName := fs.String("first-name", "", "first name (required)")
	lastName := fs.String("last-name", "", "last name (required)")
	companyID := fs.Int64("company-id", 0, "existing company ID (omit to create a new company)")
	companyName := fs.String("company-name", "My Company", "company name (used when creating a new company)")
	jobTitle := fs.String("job-title", "Admin", "job title")
	_ = fs.Parse(args)

	if *email == "" || *password == "" || *firstName == "" || *lastName == "" {
		return fmt.Errorf("--email, --password, --first-name, and --last-name are required")
	}
	if len(*password) < 8 {
		return fmt.Errorf("password must be at least 8 characters")
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	ctx := context.Background()
	pool, err := db.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()

	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	q := gen.New(tx)

	now := time.Now().UTC()
	var cid int64

	if *companyID > 0 {
		cid = *companyID
	} else {
		company, err := q.CreateCompany(ctx, gen.CreateCompanyParams{
			Name:                *companyName,
			TaxID:               "",
			Address:             "",
			CreatedAt:           now,
			Phone:               "",
			Email:               *email,
			Website:             "",
			City:                "",
			Region:              "",
			PostalCode:          "",
			Country:             "",
			Timezone:            "UTC",
			Currency:            "USD",
			SystemOfMeasurement: "imperial",
		})
		if err != nil {
			return fmt.Errorf("creating company: %w", err)
		}
		cid = company.ID

		if err := handler.SeedDefaultRoles(ctx, q, cid); err != nil {
			return fmt.Errorf("seeding default roles: %w", err)
		}
	}

	roles, err := q.ListRoles(ctx, gen.ListRolesParams{CompanyID: cid, Limit: 100, Offset: 0})
	if err != nil {
		return fmt.Errorf("listing roles: %w", err)
	}
	var roleID int64
	for _, r := range roles {
		if r.Name == "Super Admin" {
			roleID = r.ID
			break
		}
	}
	if roleID == 0 {
		return fmt.Errorf("no Super Admin role found for company %d", cid)
	}

	emp, err := q.CreateEmployee(ctx, gen.CreateEmployeeParams{
		DefaultCompanyID: &cid,
		FirstName:        *firstName,
		LastName:         *lastName,
		RoleID:           &roleID,
		IsActive:         true,
		Email:            *email,
		JobTitle:         *jobTitle,
		IsAccountOwner:   true,
		CustomFields:     []byte("{}"),
		TablePreferences: []byte("{}"),
		DashboardPreferences: []byte("{}"),
		UpdatedAt:        now,
	})
	if err != nil {
		return fmt.Errorf("creating employee: %w", err)
	}

	if err := q.AddEmployeeCompany(ctx, gen.AddEmployeeCompanyParams{
		EmployeeID: emp.ID,
		CompanyID:  cid,
	}); err != nil {
		return fmt.Errorf("linking employee to company: %w", err)
	}

	hash, err := auth.HashPassword(*password)
	if err != nil {
		return err
	}
	if err := q.UpdateEmployeePassword(ctx, gen.UpdateEmployeePasswordParams{
		ID:           emp.ID,
		PasswordHash: hash,
		UpdatedAt:    now,
	}); err != nil {
		return fmt.Errorf("setting password: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("committing transaction: %w", err)
	}

	fmt.Printf("created employee %d (%s %s) for company %d\n", emp.ID, *firstName, *lastName, cid)
	fmt.Printf("login with: POST /auth/login {email: %q, password: <your password>}\n", *email)
	return nil
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: fleet-cli <command> [flags]")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "commands:")
	fmt.Fprintln(os.Stderr, "  setpass      Set an employee's password")
	fmt.Fprintln(os.Stderr, "  createuser   Create a new employee with a company and admin role")
}
