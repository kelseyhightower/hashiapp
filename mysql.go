package main

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
)

type DBManager struct {
	DB *sql.DB
}

func NewDBManager(driverName, dbAddr string) (*DBManager, error) {
	username, password, err := getDatabaseCredentials()
	if err != nil {
		return nil, err
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s)/hashiapp", username, password, dbAddr)
	db, err := sql.Open(driverName, dsn)
	if err != nil {
		return nil, err
	}
	return &DBManager{DB: db}, nil
}

func (dbm *DBManager) Renew() {}
