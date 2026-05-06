// seed-user is a one-shot CLI that creates the first user account directly
// in the SQLite database, bypassing the registration endpoint. Useful for
// keeping registration_open=false while still being able to bootstrap.
//
// Usage:
//
//	./bin/seed-user --config config.yaml
package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/geekgonecrazy/training-log/config"
	"github.com/geekgonecrazy/training-log/core/auth"
	"github.com/geekgonecrazy/training-log/store"
	"github.com/geekgonecrazy/training-log/store/sqlite"
	"golang.org/x/term"
)

func main() {
	configPath := flag.String("config", "config.yaml", "path to config.yaml")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "load config: %v\n", err)
		os.Exit(1)
	}

	st, err := sqlite.Open(cfg.Database.Path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "open db: %v\n", err)
		os.Exit(1)
	}
	defer st.Close()

	r := bufio.NewReader(os.Stdin)
	email := strings.TrimSpace(strings.ToLower(prompt(r, "Email: ")))
	if email == "" || !strings.Contains(email, "@") {
		fmt.Fprintln(os.Stderr, "invalid email")
		os.Exit(1)
	}
	name := strings.TrimSpace(prompt(r, "Display name: "))

	password, err := promptPassword("Password: ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "read password: %v\n", err)
		os.Exit(1)
	}
	if len(password) < 8 {
		fmt.Fprintln(os.Stderr, "password must be at least 8 characters")
		os.Exit(1)
	}
	confirm, err := promptPassword("Confirm: ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "read confirm: %v\n", err)
		os.Exit(1)
	}
	if password != confirm {
		fmt.Fprintln(os.Stderr, "passwords do not match")
		os.Exit(1)
	}

	hash, err := auth.HashPassword(password)
	if err != nil {
		fmt.Fprintf(os.Stderr, "hash password: %v\n", err)
		os.Exit(1)
	}
	u, err := st.Users().Create(context.Background(), email, hash, name)
	if err != nil {
		if errors.Is(err, store.ErrConflict) {
			fmt.Fprintln(os.Stderr, "user already exists")
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "create user: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("created user %d (%s)\n", u.ID, u.Email)
}

func prompt(r *bufio.Reader, label string) string {
	fmt.Print(label)
	line, err := r.ReadString('\n')
	if err != nil {
		return ""
	}
	return line
}

func promptPassword(label string) (string, error) {
	fmt.Print(label)
	bytes, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println()
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}
