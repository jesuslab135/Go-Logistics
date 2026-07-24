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

func usage() {
	fmt.Fprintln(os.Stderr, "usage: fleet-cli setpass --email <email> --password <password>")
}
