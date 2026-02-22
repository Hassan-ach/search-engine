package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	_ "github.com/lib/pq"

	"spider/internal/config"
	"spider/internal/entity"
)

type SQLClient struct {
	conn *sql.DB
	ctx  context.Context
}

// NewDbClient creates and returns a PostgreSQL DB client.
func NewDbClient(ctx context.Context, conf *config.PSQLConfig) *SQLClient {
	psqlconn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		conf.Host,
		conf.Port,
		conf.User,
		conf.Password,
		conf.DBname,
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

	db.SetConnMaxLifetime(conf.MaxConnLifetime)
	db.SetMaxOpenConns(conf.MaxOpenConns)
	db.SetMaxIdleConns(conf.MaxIdleConns)

	fmt.Println("Data Base Connectd")

	return &SQLClient{
		conn: db, ctx: ctx,
	}
}

// Page stores a crawled page in the "pages" table.
func (c *SQLClient) InsertPage(page entity.Page) error {
	tx, err := c.conn.Begin()
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	var url_id string
	err = c.conn.QueryRowContext(c.ctx,
		`INSERT INTO urls(url) VALUES ($1) 
				ON CONFLICT (url) DO UPDATE SET url = urls.url
				RETURNING id`,
		page.URL).Scan(&url_id)
	if err != nil {
		return fmt.Errorf("upsert url: %w", err)
	}

	metadata, err := json.Marshal(page.MetaData)
	if err != nil {
		return fmt.Errorf("marshal metadata: %w", err)
	}

	_, err = c.conn.Exec(
		`INSERT INTO pages(url_id, html, metadata) VALUES ($1,$2,$3)`,
		url_id,
		page.HTML,
		metadata,
	)
	if err != nil {
		return fmt.Errorf("insert page : %w", err)
	}

	return tx.Commit()
}

// InsertGraphEdges from page id → many page ids
// - batch-inserts edges using IDs
func (c *SQLClient) InsertGraphEdges(from_url_id string, to_url_ids []string) error {
	if len(to_url_ids) == 0 {
		return nil
	}

	var (
		placeholders []string
		args         []any
	)

	for _, to_url_id := range to_url_ids {
		placeholders = append(placeholders, fmt.Sprintf("($%d, $%d)", len(args)+1, len(args)+2))
		args = append(args, from_url_id, to_url_id)
	}

	query := `INSERT INTO graph_edges (from_url, to_url) VALUES ` +
		strings.Join(placeholders, ", ") +
		` ON CONFLICT DO NOTHING`

	_, err := c.conn.ExecContext(c.ctx, query, args...)
	if err != nil {
		return fmt.Errorf("batch insert graph_edges: %w", err)
	}

	return nil
}

// InsertURLs inserts or gets existing IDs — returns []string [from_url_id, to_url_ids...]
func (c *SQLClient) InsertURLs(urls []string) ([]string, error) {
	if len(urls) == 0 {
		return make([]string, 0), nil
	}

	var values []string
	args := make([]any, 0, len(urls))

	for i, u := range urls {
		values = append(values, fmt.Sprintf("($%d)", i+1))
		args = append(args, u)
	}

	query := fmt.Sprintf(`
		WITH input(url) AS (
			SELECT DISTINCT url FROM (VALUES %s) AS v(url)
		),
		ins AS (
		    INSERT INTO urls (url)
		    SELECT url FROM input
		    ON CONFLICT (url) DO NOTHING
		    RETURNING id, url
		)
		SELECT u.id
		FROM input i
		JOIN urls u ON u.url = i.url
		`, strings.Join(values, ","))

	rows, err := c.conn.QueryContext(c.ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("batch upsert query failed: %w", err)
	}
	defer rows.Close()

	m := make([]string, 0, len(urls))

	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan failed: %w", err)
		}
		m = append(m, id)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(m) != len(urls) {
		return nil, fmt.Errorf("got %d ids back, expected %d", len(m), len(urls))
	}

	return m, nil
}
