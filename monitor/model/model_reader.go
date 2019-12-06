package model

import (
	"monitor/monitor/db"
	"reflect"
)

type ReaderContract interface {
	GetAll(models interface{})			(int64,error)

	GetOne(model interface{})			error
}

type Reader struct {
	model *modelInfo
}


func (r *Reader)GetAll(models interface{}) (int64,error) {

	mv := reflect.ValueOf(models)
	ind := reflect.Indirect(mv)

	slice := ind

	m := reflect.New(reflect.ValueOf(r.model.model).Type().Elem())


	where := []interface{}{
		[]interface{}{"id",">",22000},
	}
	rows := db.Db.Connector().Table(r.model.table).Select(r.model.fields...).Where(where...).Get()

	refs := make([]interface{},4)
	for i := range refs {
		var ref interface{}
		refs[i] = &ref
	}

	var count int64 = 0

	for rows.Next() {
		_ = rows.Scan(refs...)

		setColsValue(&m,refs...)
		slice = reflect.Append(slice,m)
		count ++
	}

	ind.Set(slice)

	return count,nil
}

func setColsValue(ind *reflect.Value,values ...interface{}) {

	fields := ind.Elem()

	for index,val := range values {
		field := fields.Field(index)
		setValue(&field,val)
	}
}

func setValue(field *reflect.Value,val interface{}){

	v := convertDataType(field,val)

	rv := reflect.ValueOf(v)

	field.Set(rv)
}

func convertDataType(field *reflect.Value,val interface{}) interface{} {
	v := reflect.Indirect(reflect.ValueOf(val)).Interface()

	var rv interface{}
	switch v := v.(type) {
	case string:
		rv = v
	case []byte:
		rv = string(v)
	case int64:
		rv = int(v)
	}

	/*switch field.Interface().(type) {
	case int:
		iv := reflect.ValueOf(val).Elem().Interface().(int64)
		return int(iv)
	case string:
		return s
	}*/

	return rv
}

func (r *Reader)GetOne(model interface{}) error {
	return nil
}


