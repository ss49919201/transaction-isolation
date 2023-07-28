package main

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/go-sql-driver/mysql"
)

// dirty read
// 他のTXの**未コミット**の更新が参照できてしまう
// 更新 => INSERT,UPDATE,DELETE
func main() {
	// DBではなくConnを生成する必要あり
	db := openDB()
	conn := newConn(db)
	conn2 := newConn(db)

	// dirty read を発生させたいセッションのみ READ UNCOMMITTED にしておく
	setTXLevel(conn)

	tx := newTx(conn)
	tx2 := newTx(conn2)

	// counterを1から2に更新
	if _, err := tx2.Query("UPDATE tbl SET counter = 2 WHERE id = 1"); err != nil {
		panic(err)
	}

	// dirty read
	var counter int
	if err := tx.QueryRow("SELECT counter FROM tbl WHERE id = 1").Scan(&counter); err != nil {
		panic(err)
	}
	fmt.Println("tx counter: ", counter) // 別のTXの未コミットの更新が参照できるので 2 が出力される

	tx.Rollback()
	tx2.Rollback()
}

func openDB() *sql.DB {
	db, err := sql.Open("mysql", dsn())
	if err != nil {
		panic(err)
	}
	return db
}

func newConn(db *sql.DB) *sql.Conn {
	conn, err := db.Conn(context.Background())
	if err != nil {
		panic(err)
	}
	return conn
}

func setTXLevel(conn *sql.Conn) {
	_, err := conn.QueryContext(context.Background(), "SET SESSION TRANSACTION ISOLATION LEVEL READ UNCOMMITTED")
	if err != nil {
		panic(err)
	}
}

func newTx(conn *sql.Conn) *sql.Tx {
	tx, err := conn.BeginTx(context.Background(), nil)
	if err != nil {
		panic(err)
	}
	return tx
}

func dsn() string {
	return (&mysql.Config{
		User:      "user",
		Passwd:    "password",
		Net:       "tcp",
		Addr:      "localhost:3306",
		DBName:    "mydb",
		ParseTime: true,
	}).FormatDSN()
}
