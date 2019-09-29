package hsql

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"strings"
	"unicode"
)

func GetContext(ctx context.Context, db *sql.DB, dest interface{}, query string, args ...interface{}) error {
	return first(ctx, db, nil, dest, query, args...)
}

func TxGetContext(ctx context.Context, tx *sql.Tx, dest interface{}, query string, args ...interface{}) error {
	return first(ctx, nil, tx, dest, query, args...)
}

func QueryContext(ctx context.Context, db *sql.DB, dest interface{}, queryStr string, args ...interface{}) error {
	return query(ctx, db, nil, dest, queryStr, args...)
}

func TxQueryContext(ctx context.Context, tx *sql.Tx, dest interface{}, queryStr string, args ...interface{}) error {
	return query(ctx, nil, tx, dest, queryStr, args...)
}

func first(ctx context.Context, db *sql.DB, tx *sql.Tx, dest interface{}, queryStr string, args ...interface{}) error {
	t := reflect.TypeOf(dest)
	v := reflect.ValueOf(dest)
	if t.Kind() != reflect.Ptr || t.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("dest invalid(%v,%v)", t.Kind(), t.Elem().Kind())
	}

	sliceType := reflect.SliceOf(t.Elem())
	dt := reflect.New(sliceType).Interface()

	var err error
	if tx != nil {
		err = query(ctx, nil, tx, dt, queryStr, args...)
	} else {
		err = query(ctx, db, tx, dt, queryStr, args...)
	}
	if err != nil {
		return err
	}

	dtVal := reflect.ValueOf(dt).Elem()
	if dtVal.Len() <= 0 {
		return sql.ErrNoRows
	}
	v.Elem().Set(dtVal.Index(0))
	return nil
}

func query(ctx context.Context, db *sql.DB, tx *sql.Tx, dest interface{}, query string, args ...interface{}) error {
	var rows *sql.Rows
	var err error
	if tx == nil {
		rows, err = db.QueryContext(ctx, query, args...)
	} else {
		rows, err = tx.QueryContext(ctx, query, args...)
	}
	if err != nil {
		return err
	}
	defer rows.Close()

	err = findAuto(rows, dest)
	return err
}

func findAuto(rows *sql.Rows, dest interface{}) error {
	columns, err := rows.Columns()
	if err != nil {
		return err
	}

	v := reflect.ValueOf(dest)
	t := v.Type()

	if t.Kind() != reflect.Ptr || t.Elem().Kind() != reflect.Slice {
		return fmt.Errorf("dest invalid(%v,%v)", t.Kind(), t.Elem().Kind())
	}

	t = t.Elem()
	results := reflect.MakeSlice(t, 0, 0)

	var isPtr bool
	if t.Elem().Kind() == reflect.Struct {
		t = t.Elem() // struct
	} else if t.Elem().Kind() == reflect.Ptr && !v.Elem().IsNil() && t.Elem().Elem().Kind() == reflect.Struct {
		isPtr = true
		t = t.Elem().Elem() // struct
	} else {
		return fmt.Errorf("dest invalid(%v,%v)", t.Elem().Kind(), t.Elem().Elem().Kind())
	}

	for rows.Next() {
		row := reflect.New(t).Elem()
		scans := getFieldAddr(columns, row)
		err := rows.Scan(scans...)
		if err != nil {
			return err
		}

		if isPtr {
			results = reflect.Append(results, row.Addr())
			continue
		}
		results = reflect.Append(results, row)
	}

	v.Elem().Set(results)
	return rows.Err()
}

func getFieldAddr(cols []string, dstVal reflect.Value) (scans []interface{}) {
	scans = make([]interface{}, len(cols))
	var noneField = sql.NullString{}
	for i, name := range cols {
		tmpName := name

		match := func(field string) bool {
			if strings.ToLower(field) == CamelToUnderscore(tmpName) {
				return true
			}

			tmpName = strings.Replace(tmpName, "_", "", -1)
			field = strings.Replace(field, "_", "", -1)
			return strings.ToLower(field) == strings.ToLower(tmpName)
		}
		fieldVal := dstVal.FieldByNameFunc(match)
		if !fieldVal.IsValid() {
			scans[i] = &noneField
		} else {
			scans[i] = fieldVal.Addr().Interface()
		}
	}
	return scans
}

func CamelToUnderscore(name string) string {
	buf := make([]rune, 0, len(name)+4)
	var preIsUpper bool
	for i, r := range name {
		if unicode.IsUpper(r) {
			if i != 0 && !preIsUpper {
				buf = append(buf, '_')
			}
			buf = append(buf, unicode.ToLower(r))
		} else {
			buf = append(buf, r)
		}
		preIsUpper = unicode.IsUpper(r)
	}
	return string(buf)
}

func UnderscoreToCamel(name string) string {
	name = strings.Replace(name, "_", " ", -1)

	name = strings.Title(name)

	return strings.Replace(name, " ", "", -1)
}
