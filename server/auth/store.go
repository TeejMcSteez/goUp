package auth

import (
	"context"
	"database/sql"
	"log"
)


func SetupStore() (*sql.DB, error) {
	db, err := sql.Open("sqlite", "./.ad.db")
	if err != nil {
		log.Printf("Error setting up auth database: %v", err)		
		return nil, err
	}
	db.SetMaxOpenConns(1)

	createTableSQL := `CREATE TABLE IF NOT EXISTS auth (
		"id" INTEGER PRIMARY KEY AUTOINCREMENET,
		"username" TEXT,
		"password" BLOB
	);`
	_, err = db.ExecContext(context.Background(), createTableSQL)
	if err != nil {
		return nil, err
	}

	log.Println("Succesfully setup auth database")
	return db, nil
}

func InsertUser(name string, password []byte, db *sql.DB) error {
	insertUserSQl := `INSERT INTO auth (username, password) VALUES (?, ?);`
	statement, err := db.Prepare(insertUserSQl)
	if err != nil {
		return err
	}
	_, err = statement.Exec(name, password)
	if err != nil {
		return err
	}
	return nil
}


