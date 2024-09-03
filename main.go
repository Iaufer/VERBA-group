package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	initdb "tdl/initDB"
	"tdl/rest"

	_ "github.com/lib/pq"
)

func createTable(db *sql.DB) {
	query := `
	CREATE TABLE IF NOT EXISTS tasks (
		id SERIAL PRIMARY KEY,
		title VARCHAR(255) NOT NULL,
		description TEXT NOT NULL,
		due_date TIMESTAMPTZ NOT NULL,
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);
	`

	_, err := db.Exec(query)
	if err != nil {
		log.Fatal("Не удалось создать таблицу:", err)
	}
}

func main() {
	password := ""
	fmt.Print("Enter a password from PostqreSQL: ")
	fmt.Scanln(&password)
	db, err := initdb.NewPostgresConnecction(initdb.ConnectionInfo{
		Host:     "localhost",
		Port:     5432,
		User:     "postgres",
		Dbname:   "postgres",
		SSLmode:  "disable",
		Password: password,
	})

	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	createTable(db)

	fmt.Println("Connected to database!")

	handler := &rest.Handler{DB: db}

	http.HandleFunc("/", handler.Handle)
	err = http.ListenAndServe(":8080", nil)

	if err != nil {
		log.Fatal(err)
	}

}
