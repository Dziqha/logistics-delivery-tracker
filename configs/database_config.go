package configs

import (
	"fmt"
	"os"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)



func DatabaseConnection() *gorm.DB {
	var (
	host     = os.Getenv("DB_HOST")
	port     = os.Getenv("DB_PORT")
	user     = os.Getenv("DB_USER")
	password = os.Getenv("DB_PASSWORD")
	dbname   = os.Getenv("DB_NAME")
	sslmode  = os.Getenv("SSL_MODE")
	timezone = os.Getenv("TIMEZONE")
	dsn      = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s TimeZone=%s",
				host, port, user, password, dbname, sslmode, timezone)
)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		fmt.Printf("Error connecting to the database: %v\n", err)
		return nil
	}

	fmt.Println("Database connection established successfully")
	db.Logger.LogMode(1) 
	return db
}