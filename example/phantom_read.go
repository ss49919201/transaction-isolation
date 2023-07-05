package main

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/go-sql-driver/mysql"
)

// ファントムリード
// TX内で１度読み出した行の集合を再度読み出した際に、他のTXの**コミット済み**の挿入が反映されて結果が変わってしまう
func main() {
	// DBではなくConnを生成する必要あり
	db, err := sql.Open("mysql", dsn())
	if err != nil {
		panic(err)
	}

	// 発生する
	{
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
		_, err = tx2.Query("INSERT INTO tbl (id, name, counter) values('2', 'B', '2')")
		if err != nil {
			panic(err)
		}

		rows, err := tx.Query("SELECT counter FROM tbl WHERE id >= 1")
		if err != nil {
			panic(err)
		}
		var count int
		for rows.Next() {
			count++
		}
		fmt.Println("tx counter: ", count) // 別のTXのコミット前の更新は反映されないので1が出力される。

		tx2.Commit()

		// non-repeatable read
		rows, err = tx.Query("SELECT counter FROM tbl WHERE id >= 1")
		if err != nil {
			panic(err)
		}
		var count2 int
		for rows.Next() {
			count2++
		}
		fmt.Println("tx counter: ", count2) // 別のTXのコミット前の更新は反映されないので1が出力される。
		tx.Rollback()
	}

	// データを戻す
	_, err = db.Query("DELETE FROM tbl WHERE id = 2")
	if err != nil {
		panic(err)
	}

	// 発生しない
	{
		conn, err := db.Conn(context.Background())
		if err != nil {
			panic(err)
		}

		conn2, err := db.Conn(context.Background())
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
		_, err = tx2.Query("INSERT INTO tbl (id, name, counter) values('2', 'B', '2')")
		if err != nil {
			panic(err)
		}

		rows, err := tx.Query("SELECT counter FROM tbl WHERE id >= 1")
		if err != nil {
			panic(err)
		}
		var count int
		for rows.Next() {
			count++
		}
		fmt.Println("tx counter: ", count) // 別のTXのコミット前の更新は反映されないので1が出力される。

		tx2.Commit()

		// non-repeatable read
		rows, err = tx.Query("SELECT counter FROM tbl WHERE id >= 1")
		if err != nil {
			panic(err)
		}
		var count2 int
		for rows.Next() {
			count2++
		}
		fmt.Println("tx counter: ", count2) // REPEATABLE READでは別のTXのコミット後の挿入は反映されないので1が出力される。
		tx.Rollback()
	}

	_, err = db.Query("DELETE FROM tbl WHERE id = 2")
	if err != nil {
		panic(err)
	}
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
