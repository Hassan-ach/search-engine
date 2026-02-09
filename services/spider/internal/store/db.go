package store

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	_ "github.com/lib/pq"

	"spider/internal/config"
	"spider/internal/entity"
	"spider/internal/utils"
)

// NewDbClient creates and returns a PostgreSQL DB client.
func NewDbClient() *sql.DB {
	psqlconn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		config.PgHost,
		config.PgPort,
		config.PgUser,
		config.PgPassword,
		config.PgDbname,
	)

	// Open connection. sql.Open does not establish a connection immediately, it validates arguments.
	db, err := sql.Open("postgres", psqlconn)
	if err != nil {
		log.Fatalf(
			"Failed to connect to postgres ERROR: %v",
			err,
		)
	}

	// Ping ensures the database is reachable and the connection is valid.
	err = db.Ping()
	if err != nil {
		log.Fatalf("Failed to ping postgres ERROR: %v", err)
	}

	fmt.Println("Data bae Connectd")

	return db
}

// Page stores a crawled page in the "pages" table.
func Page(page entity.Page) {
	// Convert the metadata struct to JSON for JSONB storage
	metadata, err := json.Marshal(page.MetaData)
	if err != nil {
		utils.Log.DB().
			Error("Failed to create metadata JSON", "error", err, "url", page.URL, "operation", "store.Page")
		return
	}

	insertStmt := `INSERT INTO "pages"("url","html","metadata") VALUES ($1,$2,$3)`
	_, err = DB.Exec(insertStmt, page.URL, page.HTML, metadata)
	if err != nil {
		utils.Log.DB().
			Error("Failed to insert page data", "error", err, "url", page.URL, "operation", "store.Page")
		return
	}

	utils.Log.DB().
		Info("Page data inserted successfully", "url", page.URL, "operation", "store.Page")
}

// AddGraphEdges source → many targets
// - upserts all URLs (source + targets)
// - gets their IDs
// - batch-inserts edges using IDs
func AddGraphEdges(source string, targets []string) {
	if len(targets) == 0 {
		return
	}

	allURLs := append([]string{source}, targets...)

	// 1. Batch upsert URLs → get id ↔ url map
	urlToID, err := upsertURLsAndGetIDs(allURLs)
	if err != nil {
		utils.Log.DB().
			Error("upsert urls failed", "error", err, "source", source, "operation", "store.AddGraphEdges")
		return
	}

	sourceID, ok := urlToID[source]
	if !ok {
		utils.Log.DB().
			Error("source url disappeared after upsert", "source", source, "operation", "store.AddGraphEdges")
		return
	}

	// 2. Prepare batch edge insert
	var (
		placeholders []string
		args         []any
	)

	for _, target := range targets {
		targetID, exists := urlToID[target]
		if !exists {
			utils.Log.DB().
				Error("target url missing after upsert", "target", target, "operation", "store.AddGraphEdges")
			return
		}

		placeholders = append(placeholders, fmt.Sprintf("($%d, $%d)", len(args)+1, len(args)+2))
		args = append(args, sourceID, targetID)
	}

	query := `INSERT INTO graph_edges (source, target) VALUES ` +
		strings.Join(placeholders, ", ") +
		` ON CONFLICT DO NOTHING` // or DO UPDATE if you have unique constraint + want to update something

	_, err = DB.Exec(query, args...)
	if err != nil {
		utils.Log.DB().
			Error("batch insert graph_edges failed", "error", err, "source", source, "operation", "store.AddGraphEdges")
		return
	}
}

// upsertURLsAndGetIDs inserts or gets existing IDs — returns map[url]→id
func upsertURLsAndGetIDs(urls []string) (map[string]string, error) {
	if len(urls) == 0 {
		return make(map[string]string), nil
	}

	// Build VALUES clause
	var values []string
	args := make([]any, 0, len(urls))

	for i, u := range urls {
		values = append(values, fmt.Sprintf("($%d)", i+1))
		args = append(args, u)
	}

	query := fmt.Sprintf(`
		WITH input(url) AS (
			VALUES %s
		),
		upserted AS (
			INSERT INTO urls (url)
			SELECT url FROM input
			ON CONFLICT (url) DO UPDATE
				SET url = EXCLUDED.url          
			RETURNING id, url
		)
		SELECT id, url FROM upserted
	`, strings.Join(values, ","))

	rows, err := DB.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("batch upsert query failed: %w", err)
	}
	defer rows.Close()

	m := make(map[string]string, len(urls))

	for rows.Next() {
		var id string
		var url string
		if err := rows.Scan(&id, &url); err != nil {
			return nil, fmt.Errorf("scan failed: %w", err)
		}
		m[url] = id
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Safety: make sure we got everything back (very rare race, but good to check)
	if len(m) != len(urls) {
		return nil, fmt.Errorf("got %d ids back, expected %d", len(m), len(urls))
	}

	return m, nil
}
