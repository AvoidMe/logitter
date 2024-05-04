package main

import (
	"database/sql"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type LogitterDB struct {
	conn *sql.DB
}

func NewDB() *LogitterDB {
	db, err := sql.Open("sqlite3", "logitter.db")
	if err != nil {
		panic(err) // TODO
	}
	result := &LogitterDB{conn: db}
	result.Init()
	return result
}

func (self *LogitterDB) Init() {
	_, err := self.conn.Exec(`
		CREATE TABLE IF NOT EXISTS records (
			id 					INTEGER NOT NULL PRIMARY KEY,
			timestamp		INTEGER NOT NULL,
			text				TEXT 		NOT NULL
		);
	`)
	if err != nil {
		panic(err) // TODO
	}
}

func (self *LogitterDB) InsertRecord(text string) {
	timestamp := time.Now().Unix()
	_, err := self.conn.Exec(`
		INSERT INTO records (timestamp, text) VALUES(?, ?);
	`,
		timestamp,
		text,
	)
	if err != nil {
		panic(err) // TODO
	}
}

func (self *LogitterDB) Close() {
	self.conn.Close()
}
