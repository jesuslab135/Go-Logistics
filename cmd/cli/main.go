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
}
