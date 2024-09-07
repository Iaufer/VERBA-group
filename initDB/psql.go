package initdb

import (
	"database/sql"
	"fmt"
)

type ConnectionInfo struct {
	Host     string
	Port     int
	User     string
	Dbname   string
	SSLmode  string
	Password string
}

func NewPostgresConnecction(info ConnectionInfo) (*sql.DB, error) {
	db, err := sql.Open("postgres", fmt.Sprintf("host=%s port=%d user=%s dbname=%s sslmode=%s password=%s", info.Host, info.Port,
		info.User, info.Dbname, info.SSLmode, info.Password))

	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}
