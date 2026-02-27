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
	GetData(words []string) (*Data, error)
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

func (s *PsqlStore) GetData(words []string) (*Data, error) {
	sql := `
		SELECT 
		    w.word, 
		    w.idf,
		    p.url_id,
			u.url,
		    pw.tf,
		    pr.score,
			p.metadata
		FROM words w
		INNER JOIN page_word pw ON w.id = pw.word_id
		INNER JOIN pages p ON pw.page_id = p.id
		INNER JOIN page_rank pr ON p.url_id = pr.url_id
		INNER JOIN urls u ON p.url_id = u.id
		WHERE w.word = ANY($1)
	`

	rows, err := s.conn.Query(sql, pq.Array(words))
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	wordIdf := make(map[string]float64, len(words))
	mapPages := make(map[uuid.UUID]*model.Page)

	for rows.Next() {
		var word string
		var idf float64
		var urlID uuid.UUID
		var url string
		var tf int
		var prScore float64
		var metadata []byte
		var meta model.MetaData

		err := rows.Scan(&word, &idf, &urlID, &url, &tf, &prScore, &metadata)
		if err != nil {
			return nil, fmt.Errorf("failed to scan data: %w", err)
		}

		err = json.Unmarshal(metadata, &meta)
		if err != nil {
			panic(fmt.Sprintf("failed to unmarshal metadata for url %s: %v", url, err))
		}

		wordIdf[word] = idf
		pg, ok := mapPages[urlID]
		if !ok {
			pg = &model.Page{
				URLID:    urlID,
				URL:      url,
				PRScore:  prScore,
				Words:    make(map[string]int, len(words)),
				MetaData: meta,
			}
			pg.MetaData = meta
			mapPages[urlID] = pg
		}
		pg.Words[word] = tf
	}

	pgs := make([]*model.Page, 0, len(mapPages))
	for _, p := range mapPages {
		pgs = append(pgs, p)
	}

	pageMapper := util.NewPageMapper()

	for _, page := range pgs {
		pageMapper.MapValue(page.URLID)
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
