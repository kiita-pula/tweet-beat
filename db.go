package main

import (
	"database/sql"
	"log"
	"os"
	"time"
)

type db struct {
	conn *sql.DB
}

func (d *db) addSubscriber(ID int) error {
	_, err := d.conn.Exec("INSERT INTO telegram_users (telegram_id, joined_at) VALUES (?,?)", ID, time.Now().Unix())
	return err
}

func (d *db) fetchSubscribers() ([]int, error) {
	rows, err := d.conn.Query("SELECT telegram_id FROM telegram_users")
	if err != nil {
		log.Println(err)
	}

	var ids []int
	var id int
	for rows.Next() {
		err := rows.Scan(&id)
		if err != nil {
			return nil, err
		}

		ids = append(ids, id)
	}

	return ids, nil
}

func initDatabase(cfg dbCfg) *db {
	f, err := os.ReadFile("subscribers.sql")
	if err != nil {
		log.Fatal(err)
	}

	conn, err := sql.Open("sqlite3", cfg.DBPath())
	if err != nil {
		log.Fatal(err)
	}

	_, err = conn.Exec(string(f))
	if err != nil {
		log.Fatal(err)
	}

	return &db{conn: conn}
}
