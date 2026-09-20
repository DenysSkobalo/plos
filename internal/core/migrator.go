package core

import (
	"database/sql"
	"embed"
	"fmt"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// RunMigrations виконує атомарні міграції з embedded FS у хронологічному порядку версій.
func RunMigrations(db *sql.DB, migrationsFS embed.FS, dir string) error {
	entries, err := migrationsFS.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("failed to read embedded migrations dir: %w", err)
	}

	type migrationFile struct {
		version int
		name    string
	}

	var files []migrationFile
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}
		parts := strings.SplitN(entry.Name(), "_", 2)
		if len(parts) < 2 {
			continue
		}
		version, err := strconv.Atoi(parts[0])
		if err != nil {
			continue
		}
		files = append(files, migrationFile{version: version, name: entry.Name()})
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].version < files[j].version
	})

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version INTEGER PRIMARY KEY,
			applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
	`)
	if err != nil {
		return fmt.Errorf("failed to create schema_migrations tracking table: %w", err)
	}

	for _, file := range files {
		var exists bool
		err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = ?)", file.version).Scan(&exists)
		if err != nil {
			return fmt.Errorf("failed to verify migration version %d: %w", file.version, err)
		}
		if exists {
			continue
		}

		content, err := migrationsFS.ReadFile(filepath.Join(dir, file.name))
		if err != nil {
			return fmt.Errorf("failed to read embedded migration file %s: %w", file.name, err)
		}

		tx, err := db.Begin()
		if err != nil {
			return fmt.Errorf("failed to begin transaction for migration %s: %w", file.name, err)
		}

		if _, err := tx.Exec(string(content)); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("failed execution of migration %s: %w", file.name, err)
		}

		if _, err := tx.Exec("INSERT INTO schema_migrations (version) VALUES (?)", file.version); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("failed to log migration version %d: %w", file.version, err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit transaction for migration %s: %w", file.name, err)
		}
	}

	return nil
}
