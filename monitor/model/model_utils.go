package model

import (
	"reflect"
	"unicode"
)

func getTableName(val reflect.Value) string {
	if method := val.MethodByName("TableName"); method.IsValid() {
		table := method.Call([]reflect.Value{})

		if len(table) > 0 && table[0].Kind() == reflect.String {
			return table[0].String()
		}
	}

	return reflect.Indirect(val).Type().Name()
}

func newModelInfo(val reflect.Value) *modelInfo {
	mi := &modelInfo{}
	mi.fields = make([]string,0)
	mi.modelName = getFullName(val)
	mi.connection = getConnection(val) // 使用默认的连接
	addFields(mi,val)
	return mi
}

func getFullName(val reflect.Value) string {
	v := reflect.Indirect(val).Type()

	return v.PkgPath() + v.Name()
}

func getConnection(val reflect.Value) string {
	if method := val.MethodByName("Connection"); method.IsValid() {
		connection := method.Call([]reflect.Value{})

		if len(connection) > 0 && connection[0].Kind() == reflect.String {
			return connection[0].String()
		}
	}

	return ""
}

func addFields(model *modelInfo,val reflect.Value) {
	vt := reflect.Indirect(val).Type()
	for i:=0 ; i < vt.NumField(); i++ {
		model.fields = append(model.fields,LcFirst(vt.Field(i).Name))
	}
}

func LcFirst(str string) string {
	for i,v := range str {
		return string(unicode.ToLower(v))+str[i+1:]
	}

	return ""
}
