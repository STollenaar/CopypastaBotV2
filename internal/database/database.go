package database

import (
	"crypto/sha256"
	"database/sql"
	"embed"
	"encoding/hex"
	"fmt"
	"log"
	"log/slog"
	"path/filepath"
	"sort"
	"strings"
	"time"

	_ "github.com/marcboeker/go-duckdb/v2" // DuckDB Go driver
	"github.com/stollenaar/copypastabotv2/internal/util"
)

var (
	duckdbClient *sql.DB

	//go:embed changelog/*.sql
	changeLogFiles embed.FS
)

func init() {

	var err error

	duckdbClient, err = sql.Open("duckdb", fmt.Sprintf("%s/copypastabot.db", util.ConfigFile.DUCKDB_PATH))

	if err != nil {
		log.Fatal(err)
	}

	// Ensure changelog table exists
	_, err = duckdbClient.Exec(`
	CREATE TABLE IF NOT EXISTS database_changelog (
		id INTEGER PRIMARY KEY,
		name VARCHAR NOT NULL,
		applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		checksum VARCHAR,
		success BOOLEAN DEFAULT TRUE
	);
	`)

	if err != nil {
		log.Fatalf("failed to create changelog table: %v", err)
	}

	if err := runMigrations(); err != nil {
		log.Fatalf("migration failed: %v", err)
	}

	slog.Info("All migrations applied successfully.")
}

func runMigrations() error {
	entries, err := changeLogFiles.ReadDir("changelog")
	if err != nil {
		return fmt.Errorf("failed to read embedded changelogs: %w", err)
	}

	var files []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".sql") {
			files = append(files, entry.Name())
		}
	}

	sort.Strings(files)

	for i, file := range files {
		id := i + 1

		contents, err := changeLogFiles.ReadFile(filepath.Join("changelog", file))
		if err != nil {
			return fmt.Errorf("failed to read %s: %w", file, err)
		}

		checksum := sha256.Sum256(contents)
		checksumHex := hex.EncodeToString(checksum[:])

		var appliedChecksum string
		err = duckdbClient.QueryRow("SELECT checksum FROM database_changelog WHERE id = ?", id).Scan(&appliedChecksum)
		if err == nil {
			if appliedChecksum != checksumHex {
				return fmt.Errorf("checksum mismatch for migration %s (id=%d). File has changed", file, id)
			}
			log.Printf("Skipping already applied migration %s", file)
			continue
		}

		// Run changelogs in a transaction
		tx, err := duckdbClient.Begin()
		if err != nil {
			return fmt.Errorf("failed to begin tx: %w", err)
		}

		_, err = tx.Exec(string(contents))
		if err != nil {
			_ = tx.Rollback()
			_, _ = duckdbClient.Exec(`
				INSERT INTO database_changelog (id, name, applied_at, checksum, success) VALUES (?, ?, ?, ?, false)
			`, id, file, time.Now(), checksumHex)
			return fmt.Errorf("failed to apply migration %s: %w", file, err)
		}

		err = tx.Commit()
		if err != nil {
			return fmt.Errorf("failed to commit migration %s: %w", file, err)
		}

		_, err = duckdbClient.Exec(`
			INSERT INTO database_changelog (id, name, applied_at, checksum, success)
			VALUES (?, ?, ?, ?, true)
		`, id, file, time.Now(), checksumHex)
		if err != nil {
			return fmt.Errorf("failed to record migration %s: %w", file, err)
		}

		log.Printf("Applied migration %s", file)
	}

	return nil
}

// ---- Speak queue --------------------------------------------------------

// QueueRecord is the persistable form of a speak queue item.
type QueueRecord struct {
	ID      int64
	GuildID string
	UserID  string
	Content string
	CmdType string
	CmdName string
	AppID   string
	Token   string
	Status  string
}

// EnqueueSpeakItem inserts a new pending item and returns its auto-assigned ID.
func EnqueueSpeakItem(r QueueRecord) (int64, error) {
	var id int64
	err := duckdbClient.QueryRow(
		`INSERT INTO speak_queue (guild_id, user_id, content, cmd_type, cmd_name, app_id, token, status)
		 VALUES (?, ?, ?, ?, ?, ?, ?, 'pending')
		 RETURNING id`,
		r.GuildID, r.UserID, r.Content, r.CmdType, r.CmdName, r.AppID, r.Token,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("enqueue speak item: %w", err)
	}
	return id, nil
}

// SetSpeakItemStatus updates the status of a queue item ("playing" or "done").
func SetSpeakItemStatus(id int64, status string) error {
	_, err := duckdbClient.Exec(`UPDATE speak_queue SET status = ? WHERE id = ?`, status, id)
	if err != nil {
		return fmt.Errorf("set speak item status: %w", err)
	}
	return nil
}

// GetPendingSpeakItems returns all items with status 'pending', ordered by creation time.
func GetPendingSpeakItems() ([]QueueRecord, error) {
	rows, err := duckdbClient.Query(
		`SELECT id, guild_id, user_id, content, cmd_type, cmd_name, app_id, token
		 FROM speak_queue WHERE status = 'pending' ORDER BY created_at`,
	)
	if err != nil {
		return nil, fmt.Errorf("get pending speak items: %w", err)
	}
	defer func() {
		err := rows.Close()
		if err != nil {
			slog.Error("Error closing rows", slog.Any("err", err))
		}
	}()

	var items []QueueRecord
	for rows.Next() {
		var r QueueRecord
		if err := rows.Scan(&r.ID, &r.GuildID, &r.UserID, &r.Content, &r.CmdType, &r.CmdName, &r.AppID, &r.Token); err != nil {
			return nil, fmt.Errorf("scanning speak item: %w", err)
		}
		items = append(items, r)
	}
	return items, rows.Err()
}
