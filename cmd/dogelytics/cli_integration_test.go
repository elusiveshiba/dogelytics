package main

import (
	"bytes"
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/dogeorg/dogelytics/internal/store"
)

func TestAdminCommandsIntegration(t *testing.T) {
	databaseURL := os.Getenv("DOGELYTICS_TEST_DBURL")
	if databaseURL == "" {
		t.Skip("DOGELYTICS_TEST_DBURL is not configured")
	}
	t.Setenv("DOGELYTICS_DBURL", databaseURL)
	email := "operator-" + time.Now().UTC().Format("20060102150405.000000000") + "@example.com"
	storage, err := store.NewStore(databaseURL, context.Background())
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	if err := storage.Migrate(context.Background()); err != nil {
		storage.Close()
		t.Fatalf("migrate test database: %v", err)
	}
	if _, err := storage.ListUsers(context.Background()); err != nil {
		storage.Close()
		t.Fatalf("check users: %v", err)
	}
	if err := storage.Close(); err != nil {
		t.Fatalf("close test database: %v", err)
	}

	var output bytes.Buffer
	if err := run(
		[]string{"admin", "user", "create", "--email", email, "--password-stdin"},
		strings.NewReader("a-long-administration-password\n"),
		&output,
		&bytes.Buffer{},
	); err != nil {
		t.Fatalf("create user: %v", err)
	}
	if !strings.Contains(output.String(), email) {
		t.Fatalf("unexpected user output: %q", output.String())
	}

	output.Reset()
	if err := run([]string{"admin", "key", "create", "--email", email}, strings.NewReader(""), &output, &bytes.Buffer{}); err != nil {
		t.Fatalf("create key: %v", err)
	}
	token := strings.TrimSpace(output.String())
	if !strings.HasPrefix(token, "dglk_") || !strings.Contains(token, ".") {
		t.Fatalf("unexpected key token: %q", token)
	}
	kid := strings.SplitN(strings.TrimPrefix(token, "dglk_"), ".", 2)[0]

	output.Reset()
	if err := run([]string{"admin", "key", "list", "--email", email}, strings.NewReader(""), &output, &bytes.Buffer{}); err != nil {
		t.Fatalf("list keys: %v", err)
	}
	if !strings.Contains(output.String(), kid+"\t"+email+"\tactive") || strings.Contains(output.String(), token) {
		t.Fatalf("unexpected key listing: %q", output.String())
	}

	output.Reset()
	if err := run([]string{"admin", "key", "revoke", "--kid", kid}, strings.NewReader(""), &output, &bytes.Buffer{}); err != nil {
		t.Fatalf("revoke key: %v", err)
	}
}
