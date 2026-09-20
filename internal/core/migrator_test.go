package core

import (
	"embed"
	"os"
	"testing"
)

//go:embed testdata/migrations/*.sql
var testMigrationsFS embed.FS

func TestRunMigrations_IdempotencyAndDataIntegrity(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test_migrator_*.db")
	if err != nil {
		t.Fatalf("failed to create temp db file: %v", err)
	}
	defer func() { _ = os.Remove(tmpFile.Name()) }()
	_ = tmpFile.Close()

	db, err := InitDB(tmpFile.Name())
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}
	defer func() { _ = db.Close() }()

	if err := RunMigrations(db, testMigrationsFS, "testdata/migrations"); err != nil {
		t.Fatalf("First RunMigrations run failed: %v", err)
	}

	if err := RunMigrations(db, testMigrationsFS, "testdata/migrations"); err != nil {
		t.Fatalf("Second RunMigrations run (idempotency check) failed: %v", err)
	}

	var debtCount int
	if err := db.QueryRow("SELECT COUNT(*) FROM debt_metadata;").Scan(&debtCount); err != nil {
		t.Fatalf("query debt_metadata count failed: %v", err)
	}
	if debtCount != 0 {
		t.Errorf("expected 0 debt entries in clean schema, got %d", debtCount)
	}

	_, err = db.Exec("INSERT INTO accounts (id, name, type, currency) VALUES ('acc-err', 'Fail Acc', 'DEBT', 'INVALID');")
	if err == nil {
		t.Errorf("expected foreign key constraint violation error, got nil")
	}
}
