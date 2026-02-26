package internal

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/lib/pq"
)

func dbSetup() *sql.DB {
	err := godotenv.Load("../../.env")
	port, err := strconv.Atoi(os.Getenv("PG_PORT"))
	if err != nil {
		panic("invalid port number")
	}

	conn, err := sql.Open("postgres", fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("PG_HOST"),
		port,
		os.Getenv("PG_USER"),
		os.Getenv("PG_PASSWORD"),
		os.Getenv("PG_DBNAME")))
	if err != nil {
		panic("failed to connect to database")
	}

	err = conn.Ping()
	if err != nil {
		log.Fatalln("")
	}

	conn.SetConnMaxLifetime(0)
	conn.SetMaxOpenConns(20)
	conn.SetMaxIdleConns(20)

	return conn
}

func GetData(conn *sql.DB, words []string) (*Data, error) {
	sql := `
		SELECT 
		    w.word, 
		    w.idf,
		    p.url_id,
			u.url,
		    pw.tf,
		    pr.score
		FROM words w
		LEFT JOIN page_word pw ON w.id = pw.word_id
		LEFT JOIN pages p ON pw.page_id = p.id
		LEFT JOIN page_rank pr ON p.url_id = pr.url_id
		LEFT JOIN urls u ON p.url_id = u.id
		WHERE w.word = ANY($1)
	`

	rows, err := conn.Query(sql, pq.Array(words))
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	wordIdf := make(map[string]float64, len(words))
	mapPages := make(map[uuid.UUID]*Page)

	for rows.Next() {
		var word string
		var idf float64
		var urlID uuid.UUID
		var url string
		var tf int
		var prScore float64

		err := rows.Scan(&word, &idf, &urlID, &url, &tf, &prScore)
		if err != nil {
			return nil, err
		}

		wordIdf[word] = idf
		pg, ok := mapPages[urlID]
		if !ok {
			pg = &Page{
				URLID:   urlID,
				URL:     url,
				PRScore: prScore,
				Words:   make(map[string]int, len(words)),
			}
			mapPages[urlID] = pg
		}
		pg.Words[word] = tf
	}

	pgs := make([]*Page, 0, len(mapPages))
	for _, p := range mapPages {
		pgs = append(pgs, p)
	}
	return &Data{
		Pages: pgs,
		Idf:   wordIdf,
	}, nil
}
