package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/mail"
	"os"
	"strings"
	"time"

	"github.com/dogeorg/dogelytics/internal/config"
	"github.com/dogeorg/dogelytics/internal/credentials"
	"github.com/dogeorg/dogelytics/internal/store"
	"golang.org/x/term"
)

var (
	version   = "dev"
	commit    = "unknown"
	buildDate = "unknown"
)

var healthcheckClient = &http.Client{Timeout: 10 * time.Second}

func run(args []string, stdin io.Reader, stdout, stderr io.Writer) error {
	if len(args) == 0 || strings.HasPrefix(args[0], "-") {
		return runServer(args)
	}
	switch args[0] {
	case "serve":
		return runServer(args[1:])
	case "version":
		if len(args) != 1 {
			return errors.New("usage: dogelytics version")
		}
		_, err := fmt.Fprintf(stdout, "dogelytics %s (commit %s, built %s)\n", version, commit, buildDate)
		return err
	case "healthcheck":
		return runHealthcheck(args[1:], stdout)
	case "admin":
		return runAdmin(args[1:], stdin, stdout, stderr)
	default:
		return fmt.Errorf("unknown command %q; expected serve, admin, healthcheck, or version", args[0])
	}
}

func runHealthcheck(args []string, stdout io.Writer) error {
	flags := flag.NewFlagSet("healthcheck", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	endpoint := flags.String("url", "http://127.0.0.1:4420/readyz", "readiness endpoint")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() != 0 {
		return errors.New("usage: dogelytics healthcheck [--url URL]")
	}
	request, err := http.NewRequestWithContext(context.Background(), http.MethodGet, *endpoint, nil)
	if err != nil {
		return fmt.Errorf("create healthcheck request: %w", err)
	}
	response, err := healthcheckClient.Do(request)
	if err != nil {
		return fmt.Errorf("healthcheck request: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(response.Body, 1024))
		return fmt.Errorf("healthcheck returned %s: %s", response.Status, strings.TrimSpace(string(body)))
	}
	_, err = fmt.Fprintln(stdout, "ready")
	return err
}

func runAdmin(args []string, stdin io.Reader, stdout, stderr io.Writer) error {
	if len(args) < 2 {
		return errors.New("usage: dogelytics admin {user|key} {create|list|revoke} [options]")
	}
	databaseURL, err := config.LoadDatabaseURL()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	storage, err := store.NewStore(databaseURL, ctx)
	if err != nil {
		return err
	}
	defer storage.Close()
	if err := storage.Migrate(ctx); err != nil {
		return err
	}

	switch args[0] {
	case "user":
		return runUserAdmin(ctx, storage, args[1:], stdin, stdout, stderr)
	case "key":
		return runKeyAdmin(ctx, storage, args[1:], stdout)
	default:
		return fmt.Errorf("unknown admin resource %q; expected user or key", args[0])
	}
}

func runUserAdmin(ctx context.Context, storage *store.Store, args []string, stdin io.Reader, stdout, stderr io.Writer) error {
	switch args[0] {
	case "create":
		flags := flag.NewFlagSet("admin user create", flag.ContinueOnError)
		flags.SetOutput(io.Discard)
		email := flags.String("email", "", "user email address")
		passwordStdin := flags.Bool("password-stdin", false, "read password from standard input")
		if err := flags.Parse(args[1:]); err != nil {
			return err
		}
		if flags.NArg() != 0 {
			return errors.New("usage: dogelytics admin user create --email EMAIL [--password-stdin]")
		}
		normalisedEmail, err := normaliseEmail(*email)
		if err != nil {
			return err
		}
		password, err := readPassword(stdin, stderr, *passwordStdin)
		if err != nil {
			return err
		}
		if len(password) < 12 {
			return errors.New("password must be at least 12 characters")
		}
		passwordHash, err := credentials.HashPassword(password)
		if err != nil {
			return fmt.Errorf("hash password: %w", err)
		}
		id, err := credentials.GenerateID(16)
		if err != nil {
			return fmt.Errorf("generate user id: %w", err)
		}
		user, err := storage.CreateUser(ctx, id, normalisedEmail, passwordHash)
		if err != nil {
			return err
		}
		_, err = fmt.Fprintf(stdout, "created user %s (%s)\n", user.Email, user.ID)
		return err
	case "list":
		if len(args) != 1 {
			return errors.New("usage: dogelytics admin user list")
		}
		users, err := storage.ListUsers(ctx)
		if err != nil {
			return err
		}
		_, _ = fmt.Fprintln(stdout, "ID\tEMAIL\tCREATED")
		for _, user := range users {
			_, _ = fmt.Fprintf(stdout, "%s\t%s\t%s\n", user.ID, user.Email, user.CreatedAt.UTC().Format(time.RFC3339))
		}
		return nil
	default:
		return fmt.Errorf("unknown user action %q; expected create or list", args[0])
	}
}

func runKeyAdmin(ctx context.Context, storage *store.Store, args []string, stdout io.Writer) error {
	switch args[0] {
	case "create":
		flags := flag.NewFlagSet("admin key create", flag.ContinueOnError)
		flags.SetOutput(io.Discard)
		email := flags.String("email", "", "owner email address")
		expires := flags.String("expires", "", "expiry date in YYYY-MM-DD format")
		if err := flags.Parse(args[1:]); err != nil {
			return err
		}
		if flags.NArg() != 0 {
			return errors.New("usage: dogelytics admin key create --email EMAIL [--expires YYYY-MM-DD]")
		}
		normalisedEmail, err := normaliseEmail(*email)
		if err != nil {
			return err
		}
		user, _, err := storage.GetUserByEmail(ctx, normalisedEmail)
		if err != nil {
			return err
		}
		if user.ID == "" {
			return fmt.Errorf("user %s does not exist", normalisedEmail)
		}
		expiresAt, err := parseExpiry(*expires)
		if err != nil {
			return err
		}
		id, err := credentials.GenerateID(16)
		if err != nil {
			return err
		}
		kid, err := credentials.GenerateID(8)
		if err != nil {
			return err
		}
		secret, err := credentials.GenerateID(24)
		if err != nil {
			return err
		}
		if _, err := storage.CreateAPIKey(ctx, id, user.ID, kid, credentials.HashAPIKeySecret(secret), expiresAt); err != nil {
			return err
		}
		_, err = fmt.Fprintf(stdout, "dglk_%s.%s\n", kid, secret)
		return err
	case "list":
		flags := flag.NewFlagSet("admin key list", flag.ContinueOnError)
		flags.SetOutput(io.Discard)
		email := flags.String("email", "", "restrict results to an owner")
		if err := flags.Parse(args[1:]); err != nil {
			return err
		}
		if flags.NArg() != 0 {
			return errors.New("usage: dogelytics admin key list [--email EMAIL]")
		}
		normalisedEmail := ""
		var err error
		if *email != "" {
			normalisedEmail, err = normaliseEmail(*email)
			if err != nil {
				return err
			}
		}
		keys, err := storage.ListAPIKeys(ctx, normalisedEmail)
		if err != nil {
			return err
		}
		_, _ = fmt.Fprintln(stdout, "KID\tEMAIL\tSTATUS\tEXPIRES\tCREATED")
		for _, key := range keys {
			status := "active"
			if key.RevokedAt.Valid {
				status = "revoked"
			} else if key.ExpiresAt.Valid && !key.ExpiresAt.Time.After(time.Now()) {
				status = "expired"
			}
			expiry := "-"
			if key.ExpiresAt.Valid {
				expiry = key.ExpiresAt.Time.UTC().Format(time.RFC3339)
			}
			_, _ = fmt.Fprintf(stdout, "%s\t%s\t%s\t%s\t%s\n", key.KID, key.Email, status, expiry, key.CreatedAt.UTC().Format(time.RFC3339))
		}
		return nil
	case "revoke":
		flags := flag.NewFlagSet("admin key revoke", flag.ContinueOnError)
		flags.SetOutput(io.Discard)
		kid := flags.String("kid", "", "key identifier")
		if err := flags.Parse(args[1:]); err != nil {
			return err
		}
		if flags.NArg() != 0 {
			return errors.New("usage: dogelytics admin key revoke --kid KID")
		}
		if strings.TrimSpace(*kid) == "" {
			return errors.New("--kid is required")
		}
		revoked, err := storage.RevokeAPIKeyByKID(ctx, strings.TrimSpace(*kid), time.Now().UTC())
		if err != nil {
			return err
		}
		if !revoked {
			return fmt.Errorf("active API key %s does not exist", *kid)
		}
		_, err = fmt.Fprintf(stdout, "revoked key %s\n", *kid)
		return err
	default:
		return fmt.Errorf("unknown key action %q; expected create, list, or revoke", args[0])
	}
}

func normaliseEmail(value string) (string, error) {
	normalised := strings.ToLower(strings.TrimSpace(value))
	address, err := mail.ParseAddress(normalised)
	if err != nil || address.Address != normalised {
		return "", errors.New("a valid --email address is required")
	}
	return normalised, nil
}

func parseExpiry(value string) (*time.Time, error) {
	if strings.TrimSpace(value) == "" {
		return nil, nil
	}
	expiresAt, err := time.Parse("2006-01-02", value)
	if err != nil {
		return nil, errors.New("--expires must use YYYY-MM-DD")
	}
	expiresAt = expiresAt.UTC()
	if !expiresAt.After(time.Now()) {
		return nil, errors.New("--expires must be in the future")
	}
	return &expiresAt, nil
}

func readPassword(stdin io.Reader, stderr io.Writer, forceStdin bool) (string, error) {
	if file, ok := stdin.(*os.File); ok && term.IsTerminal(int(file.Fd())) && !forceStdin {
		_, _ = fmt.Fprint(stderr, "Password: ")
		password, err := term.ReadPassword(int(file.Fd()))
		_, _ = fmt.Fprintln(stderr)
		if err != nil {
			return "", fmt.Errorf("read password: %w", err)
		}
		return strings.TrimSpace(string(password)), nil
	}
	reader := bufio.NewReader(io.LimitReader(stdin, 4097))
	password, err := reader.ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return "", fmt.Errorf("read password from standard input: %w", err)
	}
	if len(password) > 4096 {
		return "", errors.New("password input is too large")
	}
	return strings.TrimSpace(password), nil
}
