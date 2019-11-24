package db

import (
	. "database/sql"
	"fmt"
	"log"
	"reflect"
	"sync"
)

const DriverName = "mysql"

type MysqlPool struct {
	sync.RWMutex

	head	int
	//	当前链接池的链接数
	num 	int
	// 	可容纳的链接数
	space	int
	pool	[]*Mysql

}

type MysqlServer struct {
	// 链接池
	pool 	*MysqlPool
}

type Mysql struct {
	connector 			*DB
	connections 		map[string]*DB
	currTable 			string
	fields				string
	sql					string
	where				string
	result 				[]interface{}
}

func NewMysqlPool() *MysqlPool {
	return &MysqlPool{
		head:    0,
		num:   0,
		space: 10,
		pool:    nil,
	}
}

func NewMysqlServer() *MysqlServer {
	return &MysqlServer{
		pool: NewMysqlPool(),
	}
}

func NewMysql() *Mysql{
	return &Mysql{
		connections: make(map[string]*DB),
	}
}

func (ms *MysqlServer) GetHandle() *Mysql {
	var m *Mysql
	ms.pool.RLock()
	m = ms.pool.pop()
	ms.pool.RUnlock()
	return m
}

func (ms *MysqlServer) CheckDriverName(connection string) bool {
	localConn := connections[connection]
	if localConn["driver"] == DriverName {
		return true
	}
	return false
}

// Connection : 获取连接
func (ms *MysqlServer) Connection(connection string) BaseDbContract {
	if ok := ms.CheckDriverName(connection); !ok {
		log.Panic("Drivers Miss Match!")
	}
	m := ms.GetHandle()
	// 检测是否已经链接
	if conn, ok := m.connections[connection]; ok {
		m.connector = conn
	}
	dataSource := Db.getDataSource(connection)
	db ,err := Open(DriverName,dataSource)
	if err != nil {
		log.Panic(err.Error())
	}
	m.connections[connection] = db
	m.connector = db
	return m
}

func (ms *MysqlServer) SetHandle() bool {
	ms.pool.Lock()
	ok := ms.pool.push()
	ms.pool.Unlock()
	return ok
}

// pop : 出队列
func (pool *MysqlPool) pop() *Mysql {
	var m *Mysql
	if pool.num == 0 {
		m = NewMysql()
		pool.num++
		pool.pool = append(pool.pool,m)
	}

	if pool.head < pool.num {
		m = pool.pool[pool.head]
		pool.head++
		return m
	}

	if pool.num < pool.space {
		m = NewMysql()
		pool.num++
		pool.pool = append(pool.pool,m)
		pool.head++
		return m
	}
	return nil
}

// push : 入队列
func (pool *MysqlPool) push() bool {
	if pool.head == 0 || pool.num == 0{
		return false
	}
	pool.head--

	return true
}


func (ms *MysqlServer) registerDb(){
	Db.dbServers[DriverName] = ms
}

func (m *Mysql) Table(table string) BaseDbContract {
	m.currTable = table
	return m
}

func (m *Mysql) GetTable() string {
	return m.currTable
}

// Select: select field
func (m *Mysql) Select(fields ...interface{}) BaseDbContract {
	if len(fields) == 0 {
		m.fields = "*"
	}else {
		for _,field := range fields {
			var f = fmt.Sprintf("`%v`",field)
			if len(m.fields) != 0 {
				m.fields += ","
			}
			m.fields += f
		}
	}
	return m
}

// Where: 设置where查询条件
func (m *Mysql) Where(where ...interface{}) BaseDbContract {
	for _,w := range where {
		if len(m.where) != 0 {
			m.where += " and "
		}
		v := w.([]interface{})
		m.where += "`"+v[0].(string)+"`"
		if len(v) == 2 {
			m.where += "="
			switch d := v[1].(type) {
			case string:
				m.where += "'"+d+"'"
			case int:
				m.where += fmt.Sprintf("%v",d)
			}
		}
		if len(v) == 3 {
			m.where += v[1].(string)
			switch d := v[2].(type) {
			case string:
				m.where += "'"+d+"'"
			case int:
				m.where += fmt.Sprintf("%v",d)
			}
		}
	}
	return m
}

func (m *Mysql) Get() *Rows {
	m.sql = "SELECT "+m.fields+" FROM "+m.currTable
	if len(m.where) > 0 {
		m.sql += " WHERE "+m.where
	}
	fmt.Println(m.sql)
	stmt,err := m.connector.Prepare(m.sql)
	if err != nil {
		log.Panic(err.Error())
	}
	rows, err := stmt.Query()
	if err != nil {
		log.Panic(err.Error())
	}
	m.free()
	return rows
}

func (m *Mysql) GetOne() *Row {
	m.sql = "SELECT "+m.fields+" FROM "+m.currTable+" WHERE "+m.where
	stmt,err := m.connector.Prepare(m.sql)
	if err != nil {
		log.Panic(err.Error())
	}
	row := stmt.QueryRow()
	m.free()
	return row
}

func (m *Mysql) free() bool {
	server := reflect.ValueOf(Db.dbServers[DriverName]).Interface().(*MysqlServer)

	server.SetHandle()

	return true

}
