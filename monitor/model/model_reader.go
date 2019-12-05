package model

import (
	"fmt"
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
	fmt.Println(reflect.Append(slice,m))


	where := []interface{}{
		[]interface{}{"id",">",10},
	}
	rows := db.Db.Connector().Table(r.model.table).Select(r.model.fields...).Where(where...).Get()

	refs := make([]interface{},3)
	for i := range refs {
		var ref interface{}
		refs[i] = &ref
	}
	for rows.Next() {
		_ = rows.Scan(refs...)
		t := reflect.New(m.Elem().FieldByName("id").Type()).Elem()
		t.Set(reflect.ValueOf(refs[0]).Elem())
		fmt.Println(t)
		//fmt.Println(reflect.ValueOf(refs[0]).Elem())
		fmt.Println(reflect.Indirect(reflect.ValueOf(refs[0])).Interface())
	}

	return 0,nil
}

func (r *Reader)GetOne(model interface{}) error {
	return nil
}


