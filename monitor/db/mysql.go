package db

import (
	"database/sql"
	"log"
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
	connector 			*sql.DB
	connections 		map[string]*sql.DB
	currTable 			string
	fields				string
	sql					string
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
		connections: make(map[string]*sql.DB),
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
	db ,err := sql.Open(DriverName,dataSource)
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

func (m *Mysql) Select() BaseDbContract {
	m.fields = "id,name,mobile,createtime,updatetime"
	m.sql = "select "+m.fields+ " from "+m.currTable
	return m
}
