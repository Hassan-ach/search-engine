package store

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"

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
