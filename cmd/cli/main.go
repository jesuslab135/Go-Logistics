package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"

	"fleet/internal/auth"
	"fleet/internal/bootstrap"
	"fleet/internal/config"
	"fleet/internal/db"
	"fleet/internal/db/gen"
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
	case "bootstrap":
		err = bootstrapCompany(os.Args[2:])
	case "platform-admin":
		err = platformAdmin(os.Args[2:])
	case "inventory-drift":
		err = inventoryDrift(os.Args[2:])
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

// bootstrapCompany seeds a company's default admin role, work order statuses and
// asset statuses, and links an employee to it as an admin. It replaces Django's
// seed_asset_statuses / seed_work_order_statuses management commands, and is the
// way to repair accounts created before /companies did this transactionally —
// notably an employee with no employee_companies row, who is now denied by the
// membership check on every route.
func bootstrapCompany(args []string) error {
	fs := flag.NewFlagSet("bootstrap", flag.ExitOnError)
	email := fs.String("email", "", "employee to link as admin")
	companyID := fs.Int64("company-id", 0, "company to seed (default: the employee's default company)")
	_ = fs.Parse(args)

	if *email == "" {
		return fmt.Errorf("--email is required")
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

	target := *companyID
	if target == 0 {
		if emp.DefaultCompanyID == nil {
			return fmt.Errorf("employee %q has no default company; pass --company-id", *email)
		}
		target = *emp.DefaultCompanyID
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if err := bootstrap.Company(ctx, q.WithTx(tx), target, emp.ID); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return err
	}

	fmt.Printf("bootstrapped company %d and linked employee %d (%s) as admin\n", target, emp.ID, *email)
	return nil
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage:")
	fmt.Fprintln(os.Stderr, "  fleet-cli setpass   --email <email> --password <password>")
	fmt.Fprintln(os.Stderr, "  fleet-cli bootstrap --email <email> [--company-id <id>]")
	fmt.Fprintln(os.Stderr, "  fleet-cli platform-admin --email <email> [--revoke] | --list")
	fmt.Fprintln(os.Stderr, "  fleet-cli platform-admin --create --email <email> --first-name <name> --last-name <name> --password <password>")
	fmt.Fprintln(os.Stderr, "  fleet-cli inventory-drift")
}

// platformAdmin grants or revokes the flag that opens /api/v1/admin/*. It is a
// CLI act on purpose: the population is a handful of internal staff, and an API
// route that hands out cross-tenant access would be an attack surface with no
// user. Nobody holds the flag after the migration that added it, so the first
// grant has to happen here.
func platformAdmin(args []string) error {
	fs := flag.NewFlagSet("platform-admin", flag.ExitOnError)
	email := fs.String("email", "", "employee email")
	revoke := fs.Bool("revoke", false, "revoke instead of grant")
	list := fs.Bool("list", false, "list current platform administrators")
	create := fs.Bool("create", false, "create a dedicated platform administrator with no account and no company")
	firstName := fs.String("first-name", "", "first name (with --create)")
	lastName := fs.String("last-name", "", "last name (with --create)")
	password := fs.String("password", "", "password, at least 12 characters (with --create)")
	_ = fs.Parse(args)

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

	if *list {
		admins, err := q.ListPlatformAdmins(ctx)
		if err != nil {
			return err
		}
		if len(admins) == 0 {
			fmt.Println("no platform administrators; /api/v1/admin/* is closed to everyone")
			return nil
		}
		for _, a := range admins {
			fmt.Printf("%d\t%s %s\t%s\n", a.ID, a.FirstName, a.LastName, a.Email)
		}
		return nil
	}

	if *email == "" {
		return fmt.Errorf("--email is required (or --list)")
	}

	if *create {
		return createPlatformAdmin(ctx, q, *email, *firstName, *lastName, *password)
	}

	emp, err := q.FindEmployeeByEmail(ctx, *email)
	if err != nil {
		return fmt.Errorf("no employee for email %q: %w", *email, err)
	}

	updated, err := q.SetEmployeePlatformAdmin(ctx, gen.SetEmployeePlatformAdminParams{
		ID:              emp.ID,
		IsPlatformAdmin: !*revoke,
		UpdatedAt:       time.Now().UTC(),
	})
	if err != nil {
		return err
	}

	verb := "granted to"
	if *revoke {
		verb = "revoked from"
	}
	fmt.Printf("platform administrator %s employee %d (%s)\n", verb, updated.ID, updated.Email)
	return nil
}

// minPlatformAdminPasswordLength matches the floor for account owners: this
// login reaches every client's data, so it gets at least the same.
const minPlatformAdminPasswordLength = 12

// createPlatformAdmin creates a dedicated platform administrator: an employee
// with no account and no company, who logs in with a company-less session that
// reaches /api/v1/admin/* and nothing tenant-scoped. Use it instead of granting
// the flag to a client's own employee, which mixes the two roles.
func createPlatformAdmin(ctx context.Context, q *gen.Queries, email, firstName, lastName, password string) error {
	if firstName == "" || lastName == "" || password == "" {
		return fmt.Errorf("--create needs --first-name, --last-name and --password")
	}
	if len(password) < minPlatformAdminPasswordLength {
		return fmt.Errorf("password must be at least %d characters", minPlatformAdminPasswordLength)
	}

	hash, err := auth.HashPassword(password)
	if err != nil {
		return err
	}
	id, err := q.CreatePlatformStaffEmployee(ctx, gen.CreatePlatformStaffEmployeeParams{
		FirstName:    firstName,
		LastName:     lastName,
		Email:        email,
		PasswordHash: hash,
	})
	if err != nil {
		return fmt.Errorf("create platform administrator %q (is the email already in use?): %w", email, err)
	}

	fmt.Printf("platform administrator created: employee %d (%s), no account, no company\n", id, email)
	return nil
}

// inventoryDrift reports parts whose stock row disagrees with the sum of their
// ledger entries. Before migration 000012 the journal recorded movements that
// nothing applied, so the two diverged silently by an unknown amount.
//
// It reports rather than reconciles, deliberately. Rewriting either side to
// match the other would destroy the only evidence of what actually happened,
// and which side is right is a question about the business, not the data:
// somebody has to count the shelf.
func inventoryDrift(args []string) error {
	fs := flag.NewFlagSet("inventory-drift", flag.ExitOnError)
	_ = fs.Parse(args)

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

	rows, err := gen.New(pool).InventoryDrift(ctx)
	if err != nil {
		return err
	}
	if len(rows) == 0 {
		fmt.Println("no drift: every stock row matches the sum of its ledger entries")
		return nil
	}

	fmt.Printf("%-10s %-10s %-14s %14s %14s %14s\n", "COMPANY", "PART", "STOCK ROW", "AVAILABLE", "LEDGER SUM", "DIFFERENCE")
	for _, r := range rows {
		fmt.Printf("%-10d %-10d %-14d %14s %14s %14s\n",
			r.CompanyID, r.PartID, r.PartInventoryID,
			r.AvailableQuantity.String(), r.LedgerSum.String(),
			r.AvailableQuantity.Sub(r.LedgerSum).String())
	}
	fmt.Printf("\n%d stock rows disagree with their ledger. Nothing was changed.\n", len(rows))
	return nil
}
