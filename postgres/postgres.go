package postgres

import (
	"database/sql"
	"fmt"
	"log"

	"../config"

	// postgres driver required to connect to the db
	_ "github.com/lib/pq"
)

// Connect will connect to the postgress db given some config and returns the db interface
func Connect() *sql.DB {
	dbConfig := config.GetDbConf()
	connectionString := buildConnectionString(dbConfig)

	db, err := sql.Open("postgres", connectionString)

	if err != nil {
		log.Panic(err)
	}

	err = db.Ping()

	if err != nil {
		log.Panic(err)
	}

	return db
}

// DropSchema drops the public schema for the db
func DropSchema() {
	db := Connect()
	defer db.Close()

	if _, err := db.Exec("DROP SCHEMA public CASCADE;"); err != nil {
		panic(err)
	}

	log.Output(1, "public schema dropped")
}

// CreateSchema creates the public schema for the db
func CreateSchema() {
	db := Connect()
	defer db.Close()

	if _, err := db.Exec("CREATE SCHEMA public;"); err != nil {
		panic(err)
	}

	log.Output(1, "public schema created")
}

// Exec will execute a given SQL string on the db and return sql.Result
func Exec(sql string) (sql.Result, error) {
	db := Connect()
	defer db.Close()

	return db.Exec(sql)
}

// QueryRow will query for and return a row of data
func QueryRow(sql string) *sql.Row {
	db := Connect()
	defer db.Close()

	return db.QueryRow(sql)
}

func buildConnectionString(c *config.DbConfig) string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.HOST,
		c.PORT,
		c.USER,
		c.PASSWORD,
		c.DBNAME,
		c.SSLMODE)
}
