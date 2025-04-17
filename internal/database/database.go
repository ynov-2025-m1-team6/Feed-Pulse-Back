package database

import (
	"errors"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDatabase(user, password, dbname, host, port, sslMode string) error {
	if DB != nil {
		return nil
	}
	// https://github.com/go-gorm/postgres
	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  "host=" + host + " user=" + user + " password=" + password + " dbname=" + dbname + " port=" + port + " sslmode=" + sslMode,
		PreferSimpleProtocol: true, // disables implicit prepared statement usage
	}), &gorm.Config{})
	if err != nil {
		return errors.New("failed to connect to database, " + err.Error())
	}
	// Create uuid extension if not exists
	db.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"")
	DB = db
	return nil
}
