package hsql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"unicode"
)

var (
	tagMapperName string = "sql"

	ErrDestNil = errors.New("dest is nil")
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
	if dest == nil {
		return ErrDestNil
	}
	t := reflect.TypeOf(dest)
	v := reflect.ValueOf(dest)
	if t.Kind() != reflect.Ptr || t.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("dest is invalid(%v,%v)", t.Kind(), t.Elem().Kind())
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
	if dest == nil {
		return ErrDestNil
	}

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
		return fmt.Errorf("scan data invalid(%v,%v)", t.Kind(), t.Elem().Kind())
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
		return fmt.Errorf("scan data invalid(%v,%v)", t.Elem().Kind(), t.Elem().Elem().Kind())
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
	dstType := dstVal.Type()

	for i, name := range cols {
		scans[i] = &sql.NullString{}
		for j := 0; j < dstType.NumField(); j++ {
			if !dstVal.Field(j).CanSet() {
				continue
			}

			tag, ok := dstType.Field(j).Tag.Lookup(tagMapperName)
			if !ok {
				fieldVal := dstVal.FieldByNameFunc(func(field string) bool {
					return CamelToUnderscore(name) == strings.ToLower(field)
				})
				if fieldVal.IsValid() {
					scans[i] = fieldVal.Addr().Interface()
				}
				break
			}

			if strings.ToLower(name) == strings.ToLower(tag) {
				scans[i] = dstVal.Field(j).Addr().Interface()
				break
			}
		}
	}
	return scans
}

// 驼峰转下划线
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

// 下划线转驼峰
func UnderscoreToCamel(name string) string {
	name = strings.Replace(name, "_", " ", -1)

	name = strings.Title(name)

	return strings.Replace(name, " ", "", -1)
}
