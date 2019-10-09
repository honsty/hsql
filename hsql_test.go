package hsql

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type UserInfo struct {
	ID         int64     `sql:"id"`
	Name       string    `sql:"name"`
	Phone      string    `sql:"phone"`
	FrontCover string    `sql:"front_cover"`
	Address    string    `sql:"address"`
	Balance    int64     `sql:"balance"`
	CreatedAt  time.Time `sql:"created_at"`
	UpdatedAt  time.Time `sql:"updated_at"`
}

var (
	dsn = os.Getenv("TEST_MYSQL_DSN")
)

func getDB() *sql.DB {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		panic(err)
	}
	return db
}

// 驼峰转下划线
func TestCamelToUnderscore(t *testing.T) {
	data := []string{
		"userID",    // user_id
		"userId",    // user_id
		"uuid",      // uuid
		"UUID",      // uuid
		"UserInfo",  // user_info
		"UserAInfo", // user_ainfo
		"ABTest",    // abtest
	}
	for _, v := range data {
		s := CamelToUnderscore(v)
		t.Logf("%s->%s\n", v, s)
	}
}

// 下划线转驼峰
func TestUnderscoreToCamel(t *testing.T) {
	data := []string{
		"user_id",   // UserId
		"uuid",      // Uuid
		"id",        // Id
		"user_info", // UserInfo
		"abtest",    // Abtest
	}
	for _, v := range data {
		s := UnderscoreToCamel(v)
		t.Logf("%s->%s\n", v, s)
	}
}

func TestGetContext(t *testing.T) {
	db := getDB()
	ctx := context.Background()
	query := "SELECT * FROM users WHERE id = ?"
	for i := 1; i < 5; i++ {
		userInfo := new(UserInfo)
		err := GetContext(ctx, db, userInfo, query, i)
		switch err {
		case nil:
			t.Logf("userInfo=%#v\n", userInfo)
		case sql.ErrNoRows:
			t.Logf("err=%#v\n", err)
		default:
			t.Fatalf("get fail:%v\n", err)
		}
	}
}

func TestTxGetContext(t *testing.T) {
	ctx := context.Background()
	db := getDB()
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("err=%v\n", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	query := "SELECT * FROM users WHERE id = ?"
	for i := 1; i < 5; i++ {
		userInfo := new(UserInfo)
		err := TxGetContext(ctx, tx, userInfo, query, i)
		switch err {
		case nil:
			t.Logf("userInfo=%#v\n", userInfo)
		case sql.ErrNoRows:
			t.Logf("err=%#v\n", err)
		default:
			t.Fatalf("get fail:%v\n", err)
		}
	}
}

func TestQueryContext(t *testing.T) {
	db := getDB()
	ctx := context.Background()
	query := "SELECT * FROM users WHERE ID < ?"
	list := make([]UserInfo, 0)
	err := QueryContext(ctx, db, &list, query, 10)
	if err != nil {
		t.Fatalf("query fail:%v\n", err)
	}
	for _, v := range list {
		t.Logf("info=%#v\n", v)
	}

	list2 := make([]*UserInfo, 0)
	err = QueryContext(ctx, db, &list2, query, 10)
	if err != nil {
		t.Fatalf("query2 fail:%v\n", err)
	}
	for _, v := range list2 {
		t.Logf("info2=%#v\n", v)
	}
}

func TestTxQueryContext(t *testing.T) {
	ctx := context.Background()
	db := getDB()
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("err=%v\n", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	query := "SELECT * FROM users WHERE ID < ?"
	list := make([]UserInfo, 0)
	err = TxQueryContext(ctx, tx, &list, query, 10)
	if err != nil {
		t.Fatalf("query fail:%v\n", err)
	}
	for _, v := range list {
		t.Logf("info=%#v\n", v)
	}

	list2 := make([]*UserInfo, 0)
	err = TxQueryContext(ctx, tx, &list2, query, 10)
	if err != nil {
		t.Fatalf("query2 fail:%v\n", err)
	}
	for _, v := range list2 {
		t.Logf("info2=%#v\n", v)
	}
}
