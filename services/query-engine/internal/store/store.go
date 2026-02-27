package store

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"

	"query-engine/internal/config"
	"query-engine/internal/model"
	"query-engine/internal/util"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type Store interface {
	GetData(words []string, pageNum int) (*Data, error)
}

type Data struct {
	Pages      []*model.Page
	Idf        map[string]float64
	WordMapper util.Mapper[string]
	PageMapper util.Mapper[uuid.UUID]
}
type PsqlStore struct {
	conn *sql.DB
	conf config.StoreConfig
}

func NewStore(conf config.StoreConfig) PsqlStore {
	urls := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		conf.DB.Host, conf.DB.Port, conf.DB.User, conf.DB.Password, conf.DB.DBName)

	conn, err := sql.Open("postgres", urls)

	err = conn.Ping()
	if err != nil {
		log.Fatalln("Failed to connect to database:", err)
	}

	conn.SetMaxOpenConns(conf.DB.MaxOpenConns)
	conn.SetMaxIdleConns(conf.DB.MaxIdleConns)

	return PsqlStore{
		conn: conn,
		conf: conf,
	}
}

func (s *PsqlStore) GetData(words []string, pageNum int) (*Data, error) {
	sql := `
		WITH ranked AS (
		    SELECT
		        p.id,
		        u.url,
		        pr.score AS pr,
		        p.metadata,
		        COUNT(DISTINCT w.id) AS word_count,
		        COALESCE(
		            json_agg(
		                json_build_object(
		                    'word', w.word,
		                    'idf',  w.idf,
							'tf',	pw.tf
		                )
		            ),
		            '[]'
		        ) AS word_set
		    FROM words w
		    INNER JOIN page_word pw ON w.id = pw.word_id
		    INNER JOIN pages p      ON pw.page_id = p.id
		    INNER JOIN page_rank pr ON p.url_id = pr.url_id
		    INNER JOIN urls u       ON p.url_id = u.id
		    WHERE w.word = ANY($1)
		    GROUP BY p.id, pr.score, u.url, p.metadata
		)
		SELECT *
		FROM ranked
		ORDER BY word_count DESC,
		         pr DESC,
		         id ASC
		LIMIT $2 OFFSET $3`

	rows, err := s.conn.Query(sql, pq.Array(words), s.conf.PageSize, pageNum*s.conf.PageSize)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	wordIdf := make(map[string]float64, len(words))
	pgs := make([]*model.Page, 0, s.conf.PageSize)

	for rows.Next() {
		var (
			id         uuid.UUID
			url        string
			prScore    float64
			metadata   []byte
			word_count int
			word_set   []byte

			meta    model.MetaData
			wordSet []model.Word
		)

		err := rows.Scan(&id, &url, &prScore, &metadata, &word_count, &word_set)
		if err != nil {
			return nil, fmt.Errorf("failed to scan data: %w", err)
		}

		err = json.Unmarshal(metadata, &meta)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata for url %s: %w", url, err)
		}

		err = json.Unmarshal(word_set, &wordSet)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal word set for url %s: %w", url, err)
		}
		page := &model.Page{
			ID:       id,
			URL:      url,
			PRScore:  prScore,
			Words:    make(map[string]int, len(wordSet)),
			MetaData: meta,
		}

		for _, w := range wordSet {
			wordIdf[w.Word] = w.Idf
			page.Words[w.Word] = w.Tf
		}

		pgs = append(pgs, page)
	}

	pageMapper := util.NewPageMapper()

	for _, page := range pgs {
		pageMapper.MapValue(page.ID)
	}
	wordMapper := util.NewWordMapper()
	for _, w := range words {
		wordMapper.MapValue(w)
	}

	return &Data{
		Pages:      pgs,
		Idf:        wordIdf,
		WordMapper: wordMapper,
		PageMapper: pageMapper,
	}, nil
}
