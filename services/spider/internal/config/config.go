package config

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

var (
	PgHost     string
	PgPort     int
	PgUser     string
	PgPassword string
	PgDbname   string

	RdAddr     string
	RdPassword string
	RdDb       int
	RdPort     int

	MaxGoRoutines int
)

func init() {
	err := godotenv.Load("../../.env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	PgHost = os.Getenv("PG_HOST")
	PgPort, err = strconv.Atoi(os.Getenv("PG_PORT"))
	if err != nil {
		PgPort = 5432
	}
	PgUser = os.Getenv("PG_USER")
	PgPassword = os.Getenv("PG_PASSWORD")
	PgDbname = os.Getenv("PG_DBNAME")
	MaxGoRoutines, err = strconv.Atoi(os.Getenv("MAX_GO_ROUTINES"))
	if err != nil {
		MaxGoRoutines = 10
	}

	RdAddr = os.Getenv("REDIS_ADDR")
	RdPassword = os.Getenv("REDIS_PASSWORD")
	RdDb, err = strconv.Atoi(os.Getenv("REDIS_DB"))
	if err != nil {
		RdDb = 1
	}
	RdPort, err = strconv.Atoi(os.Getenv("REDIS_PORT"))
	if err != nil {
		RdPort = 6379
	}
	fmt.Println("=== Environment Variables ===")
	fmt.Printf("REDIS_ADDR: %s\n", os.Getenv("REDIS_ADDR"))
	fmt.Printf("REDIS_PASSWORD: %s\n", os.Getenv("REDIS_PASSWORD"))
	fmt.Printf("REDIS_DB: %s\n", os.Getenv("REDIS_DB"))
	fmt.Printf("REDIS_PORT: %s\n", os.Getenv("REDIS_PORT"))
	fmt.Printf("PG_HOST: %s\n", os.Getenv("PG_HOST"))
	fmt.Printf("PG_PORT: %s\n", os.Getenv("PG_PORT"))
	fmt.Printf("PG_USER: %s\n", os.Getenv("PG_USER"))
	fmt.Printf("PG_PASSWORD: %s\n", os.Getenv("PG_PASSWORD"))
	fmt.Printf("PG_DBNAME: %s\n", os.Getenv("PG_DBNAME"))
	fmt.Printf("MAX_GO_ROUTINES: %s\n", os.Getenv("MAX_GO_ROUTINES"))
	fmt.Println("=============================")
}
