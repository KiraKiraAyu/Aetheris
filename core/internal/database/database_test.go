package database

import (
	"strings"
	"testing"
)

func TestOpenSQLite(t *testing.T) {
	// Open SQLite in-memory database
	db, err := Open("sqlite://:memory:")
	if err != nil {
		t.Fatalf("Open(sqlite://:memory:) failed: %v", err)
	}
	if db == nil {
		t.Fatal("Open(sqlite://:memory:) returned nil db")
	}

	// Verify it's actually SQLite by querying sqlite_master
	var name string
	err = db.Raw("SELECT name FROM sqlite_master WHERE type='table'").Scan(&name).Error
	if err != nil {
		t.Fatalf("Failed to execute SQLite query: %v", err)
	}
}

func TestOpenPostgres(t *testing.T) {
	// Open Postgres connection (should fail because server is down, but should use postgres driver)
	// We use an invalid/down port to fail fast.
	_, err := Open("postgres://postgres:postgres@localhost:54321/nonexistent?sslmode=disable&connect_timeout=1")
	if err == nil {
		t.Fatal("Expected Postgres Open to fail, but it succeeded")
	}

	// The error message should indicate connection failure, not "not implemented"
	if strings.Contains(err.Error(), "not implemented") {
		t.Fatalf("Expected Postgres driver to be used, but got error: %v", err)
	}
}
