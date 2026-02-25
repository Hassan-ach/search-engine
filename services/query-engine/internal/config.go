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

type RankerConfig struct {
	maxResults int
	weightTF   float64
}

type Config struct {
	db     DBConfig
	ranker RankerConfig
}
