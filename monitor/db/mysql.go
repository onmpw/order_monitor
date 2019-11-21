package db

import (
	"database/sql"
	"log"
)

type MysqlServer struct {
	//	当前链接池的链接数
	num 	int32
	// 	可容纳的链接数
	space	int32
}

type Mysql struct {
	connector 		*sql.DB
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
	Db.dbServers[m.DriverName] = m
}

func (m *Mysql) Connection(connection string) BaseDbContract {
	if ok := m.CheckDriverName(connection); !ok {
		log.Panic("Drivers Miss Match!")
	}
	// 检测是否已经链接
	if conn, ok := m.connections[connection]; ok {
		m.connector = conn
	}
	dataSource := Db.getDataSource(connection)
	db ,err := sql.Open(m.DriverName,dataSource)
	if err != nil {
		log.Panic(err.Error())
	}
	m.connections[connection] = db
	m.connector = db
	return m
}

func (m *Mysql) Table(table string) BaseDbContract {

	return m
}

func (m *Mysql) CheckDriverName(connection string) bool {
	localConn := connections[connection]
	if localConn["driver"] == m.DriverName {
		return true
	}
	return false
}
