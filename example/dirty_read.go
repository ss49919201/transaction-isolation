package main

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/go-sql-driver/mysql"
)

// dirty read
// 他のTXの**未コミット**の更新が参照できてしまう
func main() {
	// DBではなくConnを生成する必要あり
	db, err := sql.Open("mysql", dsn())
	if err != nil {
		panic(err)
	}
	if err := ping(db); err != nil {
		panic(err)
	}
	conn, err := db.Conn(context.Background())
	if err != nil {
		panic(err)
	}

	db2, err := sql.Open("mysql", dsn())
	if err != nil {
		panic(err)
	}
	if err := ping(db2); err != nil {
		panic(err)
	}
	conn2, err := db2.Conn(context.Background())
	if err != nil {
		panic(err)
	}

	// dirty read を発生させたいセッションのみ READ UNCOMMITTED にしておく
	_, err = conn.QueryContext(context.Background(), "SET SESSION TRANSACTION ISOLATION LEVEL READ UNCOMMITTED")
	if err != nil {
		panic(err)
	}

	tx, err := conn.BeginTx(context.Background(), nil)
	if err != nil {
		panic(err)
	}

	tx2, err := conn2.BeginTx(context.Background(), nil)
	if err != nil {
		panic(err)
	}
	_, err = tx2.Query("UPDATE tbl SET counter = 2 WHERE id = 1")
	if err != nil {
		panic(err)
	}

	// dirty read
	var counter int
	err = tx.QueryRow("SELECT counter FROM tbl WHERE id = 1").Scan(&counter)
	if err != nil {
		panic(err)
	}
	fmt.Println("tx counter: ", counter) // 別のTXの未コミットの更新が参照できるので 2 が出力される

	tx.Rollback()
	tx2.Rollback()
}

func ping(db *sql.DB) error {
	return db.Ping()
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
