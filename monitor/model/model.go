package model

import (
	"fmt"
	"reflect"
)

var (
	modelContainer = &models {
		container : make(map[string]*modelInfo),
		containerByFullName:make(map[string]*modelInfo),
	}
)

type field struct {
	name 			string
}

type modelInfo struct {
	modelName		string
	model			interface{}
	fields			[]string
	connection 		string

}

type models struct {
	container 				map[string]*modelInfo
	containerByFullName		map[string]*modelInfo
}

type ContractModel interface {
	GetOne()	*modelInfo
	Get()		[]*modelInfo
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
		panic(fmt.Errorf("cannot use non-ptr model struct `%s`", getFullName(sv)))
	}

	if st.Kind() == reflect.Ptr {
		panic(fmt.Errorf("only allow ptr model struct, it looks you use two reference to the struct `%s`", st))
	}

	name := getFullName(sv)

	if _,ok := modelContainer.fetchModelByFullName(name); ok {
		panic(fmt.Errorf("model `%s` repeat register, must be unique\n", name))
	}

	table := getTableName(sv)

	if _,ok := modelContainer.fetchModelByTable(table); ok {
		panic(fmt.Errorf("table name `%s` repeat register, must be unique\n", table))
	}

	modelInfo := newModelInfo(sv)
	modelInfo.model = model

	modelContainer.add(table,modelInfo)

}

func (m *models)fetchModelByFullName(name string)(*modelInfo,bool) {
	mi,ok := m.containerByFullName[name]
	return mi, ok
}

func (m *models)fetchModelByTable(table string) (*modelInfo,bool) {
	mi,ok := m.container[table]
	return mi,ok
}

func (m *models)fetchModel(model interface{},needPtr bool)(*modelInfo,bool) {
	sv := reflect.ValueOf(model)

	if needPtr && sv.Kind() != reflect.Ptr {
		panic(fmt.Errorf("cannot use non-ptr model struct `%s`", getFullName(sv)))
	}

	name := getFullName(sv)

	if mi,ok := m.fetchModelByFullName(name); ok {
		return mi,ok
	}
	return nil,false
}

func (m *models)add(table string,model *modelInfo) bool {
	m.container[table] = model
	m.containerByFullName[model.modelName] = model

	return true
}

func Read(model interface{}) {
	mi ,ok := modelContainer.fetchModel(model,true)
	if !ok {
		panic(fmt.Errorf("model `%s` has not been registered！", reflect.Indirect(reflect.ValueOf(model)).Type().Name()))
	}

	fmt.Println(mi.modelName)
}