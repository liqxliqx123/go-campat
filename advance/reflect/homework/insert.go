package homework

import (
	"database/sql/driver"
	"errors"
	"reflect"
	"strings"
)

var errInvalidEntity = errors.New("invalid entity")

func InsertStmtCore(entity interface{}) (string, []string, []interface{}, error) {
	var cols []string
	var args []interface{}
	var tableName string

	typ := reflect.TypeOf(entity)

	if entity == nil {
		return "", nil, nil, errInvalidEntity
	}

	val := reflect.ValueOf(entity)

	if val.Kind() == reflect.Ptr {
		for val.Elem().Kind() == reflect.Ptr {
			val = val.Elem()
		}
		if val.Elem().Kind() == reflect.Struct {
			val = val.Elem()
		}
	}

	typ = val.Type()
	if typ.Kind() == reflect.Struct && typ.NumField() == 0 {
		return "", nil, nil, errInvalidEntity
	}
	tableName = typ.Name()

	for i := 0; i < typ.NumField(); i++ {
		fd := typ.Field(i)
		fv := val.Field(i)

		if fd.Type.Kind() == reflect.Struct && !(fd.Type.Implements(reflect.TypeOf((*driver.Valuer)(nil)).Elem())) {
			//fmt.Println(fd.Type.Implements(reflect.TypeOf((*sql.Scanner)(nil)).Elem()))
			_, cols2, args2, err := InsertStmtCore(fv.Interface())
			if err != nil {
				return "", nil, nil, err
			}
			cols = append(cols, cols2...)
			args = append(args, args2...)
		} else {
			cols = append(cols, "`"+fd.Name+"`")
			args = append(args, fv.Interface())
		}
	}

	return tableName, cols, args, nil
}

func InsertStmt(entity interface{}) (string, []interface{}, error) {

	tableName, cols, args, err := InsertStmtCore(entity)
	if err != nil {
		return "", nil, err
	}

	bd := strings.Builder{}
	bd.WriteString("INSERT INTO `")
	bd.WriteString(tableName)
	bd.WriteString("`")

	//sqlStr := "INSERT INTO `" + tableName + "`"
	//fmt.Println(reflect.TypeOf(entity).Name())
	placeholder := make([]string, len(args))
	for p := 0; p < len(placeholder); p++ {
		placeholder[p] = "?"
	}
	//sqlStr += "(" + strings.Join(cols, ",") + ") VALUES(" + strings.Join(placeholder, ",") + ");"
	bd.WriteString("(")
	bd.WriteString(strings.Join(cols, ","))
	bd.WriteString(") VALUES(")
	bd.WriteString(strings.Join(placeholder, ","))
	bd.WriteString(");")

	return bd.String(), args, nil
}