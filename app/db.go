package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
)

func connectToDB() (*sql.DB, error) {
	const (
		host     = "Ubuntu"
		port     = 5432
		user     = "postgres"
		password = "postgres"
		dbname   = "tag_db"
	)

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatal(err)
		return nil, errors.New(err.Error())
	}
	// defer db.Close()

	if err = db.Ping(); err != nil {
		log.Fatal(err)
		return nil, errors.New(err.Error())
	}

	return db, nil
}
