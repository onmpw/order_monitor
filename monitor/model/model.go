package model

import (
	"fmt"
	"reflect"
)

type Model struct {

}

type ContractModel interface {
	GetOne()	*Model
	Get()		[]*Model
	Count()		int64
}

func RegisterModel(models ...interface{}) {
	for _, model := range models {
		register(model)
	}
}

func register(model interface{}) {
	sv := reflect.ValueOf(model)
	st := reflect.Indirect(sv).Type()

	if sv.Kind() != reflect.Ptr {
		panic(fmt.Errorf("<orm.RegisterModel> cannot use non-ptr model struct `%s`", st.PkgPath()+"."+st.Name()))
	}

	if st.Kind() == reflect.Ptr {
		panic(fmt.Errorf("<orm.RegisterModel> only allow ptr model struct, it looks you use two reference to the struct `%s`", st))
	}

}