package main

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/go-sql-driver/mysql"
)

// non-repeatable read
// 他のTXの**未コミット**の更新が参照できてしまう
func main() {
	// DBではなくConnを生成する必要あり
	db, err := sql.Open("mysql", dsn())
	if err != nil {
		panic(err)
	}

	conn, err := db.Conn(context.Background())
	if err != nil {
		panic(err)
	}

	conn2, err := db.Conn(context.Background())
	if err != nil {
		panic(err)
	}

	// non-repeatable read を発生させたいセッションのみ READ COMMITTED にしておく
	_, err = conn.QueryContext(context.Background(), "SET SESSION TRANSACTION ISOLATION LEVEL READ COMMITTED")
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

	var counter int
	err = tx.QueryRow("SELECT counter FROM tbl WHERE id = 1").Scan(&counter)
	if err != nil {
		panic(err)
	}
	fmt.Println("tx counter: ", counter) // 別のTXのコミット前の更新は反映されないので1が出力される。

	tx2.Commit()

	// non-repeatable read
	err = tx.QueryRow("SELECT counter FROM tbl WHERE id = 1").Scan(&counter)
	if err != nil {
		panic(err)
	}
	fmt.Println("tx counter: ", counter) // 別のTXのコミット後の更新が反映されるので2が出力される。

	tx.Rollback()
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