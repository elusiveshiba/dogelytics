package main

import (
	"testing"

	"github.com/dogeorg/dogelytics/internal/config"
)

func TestValidateUIConfig(t *testing.T) {
	t.Run("admin ui requires session secret", func(t *testing.T) {
		err := validateUIConfig(&config.Config{EnableAdminUI: true})
		if err == nil {
			t.Fatal("expected error when admin UI is enabled without a session secret")
		}
	})

	t.Run("dashboard ui does not require session secret", func(t *testing.T) {
		err := validateUIConfig(&config.Config{EnableDashboardUI: true})
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})
}
