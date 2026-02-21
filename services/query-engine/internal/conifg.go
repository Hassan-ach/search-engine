package internal

type DBConfig struct {
	host     string
	port     int
	user     string
	password string
	dbName   string
	maxConns int
	maxIdle  int
}

type RankingConfig struct {
	maxResults int
	weightTF   float64
}
