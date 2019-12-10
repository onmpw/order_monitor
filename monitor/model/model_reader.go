package model

import (
	"fmt"
	"log"
	"monitor/monitor/db"
	"reflect"
	"strconv"
)

type ReaderContract interface {
	GetAll(models interface{}) (int64, error)

	GetOne(model interface{}) error

	Filter(key string, val ...interface{}) ReaderContract

	Count()						int64
}

type Reader struct {
	model 		*modelInfo
	where 		[]interface{}
}

func (r *Reader) Filter(key string, val ...interface{}) ReaderContract {

	if len(val) ==0 || len(val) > 2 {
		log.Panic(fmt.Errorf("参数 val 数量不能大于2个或者小于1个，当前参数数量为%d",len(val)))
	}

	var where []interface{}
	if len(val) == 1 {
		where = []interface{}{key,"=",val[0]}
	}else {
		where = []interface{}{key,val[0],val[1]}
	}

	r.where = append(r.where,where)
	return r
}

func (r *Reader) GetAll(models interface{}) (int64, error) {

	mv := reflect.ValueOf(models)
	ind := reflect.Indirect(mv)

	slice := ind

	connect := db.Db.Connector()
	if r.model.connection != "" {
		connect = db.Db.GetConnection(r.model.connection)
	}
	rows := connect.Table(r.model.table).Select(r.model.fields...).Where(r.where...).Get()

	refs := make([]interface{}, len(r.model.fields))
	for i := range refs {
		var ref interface{}
		refs[i] = &ref
	}


	var count int64 = 0

	for rows.Next() {
		_ = rows.Scan(refs...)

		m := reflect.New(reflect.ValueOf(r.model.model).Type().Elem())

		setColsValue(&m, refs...)
		slice = reflect.Append(slice, m)
		count++
	}

	ind.Set(slice)

	return count, nil
}

func setColsValue(ind *reflect.Value, values ...interface{}) {

	fields := ind.Elem()

	for index, val := range values {
		field := fields.Field(index)
		setValue(&field, val)
	}
}

func setValue(field *reflect.Value, val interface{}) {

	v := convertDataType(field, val)

	rv := reflect.ValueOf(v)
	field.Set(rv)
}

func convertDataType(field *reflect.Value, val interface{}) interface{} {
	vv := reflect.Indirect(reflect.ValueOf(val)).Interface()

	var t int

	var s string
	switch v := vv.(type) {
	case string:
		s = v
		t = 1
	case []byte:
		s = string(v)
		t = 1
	}
	switch field.Interface().(type) {
	case int:
		if t == 0 {
			iv := reflect.ValueOf(val).Elem().Interface().(int64)
			return int(iv)
		}else {
			iv,_ := strconv.ParseInt(s,0,0)
			return int(iv)
		}
	case string:
		return s
	}

	return 0
}

func (r *Reader) GetOne(model interface{}) error {
	mv := reflect.ValueOf(model)
	ind := reflect.Indirect(mv)

	obj := reflect.New(reflect.ValueOf(r.model.model).Type().Elem())

	row := db.Db.Connector().Table(r.model.table).Select(r.model.fields...).Where(r.where...).GetOne()

	refs := make([]interface{}, len(r.model.fields))
	for i := range refs {
		var ref interface{}
		refs[i] = &ref
	}

	err := row.Scan(refs...)

	if err != nil {
		return err
	}

	setColsValue(&obj,refs...)

	if ind.CanSet() {
		ind.Set(obj)
	}
	return nil
}

func (r *Reader) Count() int64 {

	count,err := db.Db.Connector().Table(r.model.table).Where(r.where...).Count()

	if err != nil {
		log.Panic(err.Error())
	}

	return count
}
