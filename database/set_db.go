package database

import (
	"database/sql"
	"fmt"
	"google_login/config"
	"google_login/models"
	"log"

	_ "github.com/lib/pq"
)

var DB *sql.DB

// Load ENV
func NewConnDB() *models.ConnDB {
	c := config.AppConfig.Database
	return &models.ConnDB{
		Host:         c.HOST,
		Port:         c.PORT,
		DatabaseName: c.DATABASENAME,
		Username:     c.USER_NAME,
		Password:     c.PASSWORD,
	}
}

func SetupDatabase() {
	conn := NewConnDB()

	connectionString := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		conn.Host, conn.Port, conn.Username, conn.Password, conn.DatabaseName,
	)

	var err error

	DB, err = sql.Open("postgres", connectionString)
	if err != nil {
		log.Fatal(err)
	}

	if err = DB.Ping(); err != nil {
		log.Fatal("Database Ping Error:", err)
	}

	fmt.Println("Successfully connected to Database on port", conn.Port)
}
