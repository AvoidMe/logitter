package main

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type LogitterDB struct {
	conn *sql.DB
}

type Record struct {
	ID        int64
	Timestamp int64
	Text      string
}

func NewDB() (*LogitterDB, error) {
	db, err := sql.Open("sqlite3", "logitter.db")
	if err != nil {
		return nil, err
	}
	result := &LogitterDB{conn: db}
	err = result.Init()
	if err != nil {
		db.Close()
		return nil, err
	}
	return result, nil
}

func (self *LogitterDB) Init() error {
	_, err := self.conn.Exec(`
		-- Primary table
		CREATE TABLE IF NOT EXISTS records (
			id 					INTEGER NOT NULL PRIMARY KEY,
			timestamp		INTEGER NOT NULL,
			text				TEXT 		NOT NULL
		);

		-- Text index
		CREATE VIRTUAL TABLE IF NOT EXISTS text_index USING FTS5(text);

		-- Insert trigger
		CREATE TRIGGER IF NOT EXISTS insert_text AFTER INSERT ON records BEGIN
			INSERT INTO text_index(rowid, text) VALUES (new.rowid, new.text);
		END;

		-- Update trigger
		CREATE TRIGGER IF NOT EXISTS update_text AFTER UPDATE ON records BEGIN
			DELETE FROM text_index WHERE rowid = old.rowid;
			INSERT INTO text_index(rowid, text) VALUES (new.rowid, new.text);
		END;

		-- Delete trigger
		CREATE TRIGGER IF NOT EXISTS delete_text AFTER DELETE ON records BEGIN
			DELETE FROM text_index WHERE rowid = old.rowid;
		END;
	`)
	return err
}

func (self *LogitterDB) InsertRecord(text string) error {
	timestamp := time.Now().Unix()
	_, err := self.conn.Exec(`
		INSERT INTO records (timestamp, text) VALUES(?, ?);
	`,
		timestamp,
		text,
	)
	return err
}

func (self *LogitterDB) GetRecords() ([]Record, error) {
	cursor, err := self.conn.Query(`SELECT id, timestamp, text FROM records;`)
	if err != nil {
		return nil, err
	}
	defer cursor.Close()
	result := []Record{}
	for cursor.Next() {
		record := Record{}
		err := cursor.Scan(&record.ID, &record.Timestamp, &record.Text)
		if err != nil {
			return nil, err
		}
		result = append(result, record)
	}
	return result, nil
}

func (self *LogitterDB) GetRecordsFilter(text string) ([]Record, error) {
	if len(text) == 0 {
		return self.GetRecords()
	}
	cursor, err := self.conn.Query(
		`SELECT rowid FROM text_index WHERE text MATCH ?;`,
		fmt.Sprintf("%s OR %s*", text, text),
	)
	if err != nil {
		return nil, err
	}
	defer cursor.Close()

	var sb strings.Builder
	sb.WriteString("SELECT * FROM records WHERE id in (")
	ids_map := make(map[int]struct{})
	for cursor.Next() {
		var id int
		err = cursor.Scan(&id)
		if err != nil {
			return nil, err
		}
		ids_map[id] = struct{}{}
	}
	cursor.Close()

	cursor, err = self.conn.Query(
		`SELECT id FROM records WHERE text LIKE ?;`,
		fmt.Sprintf("%%%s%%", text),
	)
	if err != nil {
		return nil, err
	}
	defer cursor.Close()
	for cursor.Next() {
		var id int
		err = cursor.Scan(&id)
		if err != nil {
			return nil, err
		}
		ids_map[id] = struct{}{}
	}
	cursor.Close()

	if len(ids_map) == 0 {
		return nil, nil
	}

	var ids []any
	for id := range ids_map {
		if len(ids) > 0 {
			sb.WriteString(",")
		}
		sb.WriteString("?")
		ids = append(ids, id)
	}
	sb.WriteString(");")

	cursor, err = self.conn.Query(sb.String(), ids...)
	if err != nil {
		return nil, err
	}
	defer cursor.Close()
	result := []Record{}
	for cursor.Next() {
		record := Record{}
		err := cursor.Scan(&record.ID, &record.Timestamp, &record.Text)
		if err != nil {
			return nil, err
		}
		result = append(result, record)
	}
	return result, nil
}

func (self *LogitterDB) Close() {
	self.conn.Close()
}
