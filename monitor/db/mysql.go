package db

import (
	"database/sql"
	"log"
)

type Mysql struct {
	DriverName 		string
	connections map[string]*sql.DB
}

func NewMysql() *Mysql{
	return &Mysql{
		DriverName: "mysql",
		connections: make(map[string]*sql.DB),
	}
}

func (m *Mysql) registerDb(){
	Db.dbContainer[m.DriverName] = m
}

func (m *Mysql) Connection(connection string) *sql.DB {
	// 检测是否已经链接
	if conn, ok := m.connections[connection]; ok {
		return conn
	}
	db ,err := sql.Open(m.DriverName,Db.getDataSource(connection))
	if err != nil {
		log.Panic(err.Error())
	}
	m.connections[connection] = db
	return db
}
