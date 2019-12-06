package db

import (
	"context"
	. "database/sql"
	"fmt"
	"log"
	"reflect"
	"sync"
	"time"
)

const DriverName = "mysql"

type MysqlPool struct {
	sync.RWMutex

	// 指向队列头
	head	int
	// 指向队列尾
	tail	int
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
	stmt				map[string]*Stmt
	Err 				error
}

func NewMysqlPool() *MysqlPool {
	return &MysqlPool{
		head:    0,
		tail:	0,
		num:   0,
		space: 15,
		pool: make([]*Mysql,0),
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
		stmt: make(map[string]*Stmt),
	}
}

func (ms *MysqlServer) GetHandle() *Mysql {
	var mysql *Mysql

	// 设定10秒 如果超过10秒还没有其他的Mysql连接对象被释放，则返回空
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	select {
	case m:= <- ms.pool.pop():
		mysql = m
		break
	case <-ctx.Done():
		mysql = nil
		break
	}
	return mysql
}

func (ms *MysqlServer) SetHandle(m *Mysql) bool {
	ms.pool.Lock()
	ok := ms.pool.push(m)
	ms.pool.Unlock()
	return ok
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
		return m
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

// pop : 出队列
func (pool *MysqlPool) pop() <- chan *Mysql {
	var m *Mysql
	var mysql = make(chan *Mysql,1)
	for pool.num == pool.space && pool.head == pool.tail {} // 如果队列已经满了，并且队列中的所有元素都正在使用，则等待释放元素
	pool.RLock()
	if pool.num == 0 {
		m = NewMysql()
		pool.num++
		pool.RUnlock()
		mysql <- m
		return mysql
	}

	if pool.head < pool.tail {
		m = pool.pool[pool.head]
		pool.pool = append(pool.pool[:pool.head],pool.pool[pool.head+1:]...)
		pool.tail = len(pool.pool)
		pool.RUnlock()
		mysql <- m
		return mysql
	}

	if pool.num < pool.space {
		m = NewMysql()
		pool.num++
		pool.RUnlock()
		mysql <- m
		return mysql
	}
	pool.RUnlock()
	return nil
}

// push : 入队列
func (pool *MysqlPool) push(m *Mysql) bool {
	if pool.num == 0{
		return false
	}
	pool.pool = append(pool.pool,m)
	pool.tail = len(pool.pool)

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

// Select: 设置要查找的字段
func (m *Mysql) Select(fields ...string) BaseDbContract {
	if len(fields) == 0 {
		m.fields = "*"
	}else {
		for _,field := range fields {
			var f = fmt.Sprintf("%v",field)
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

// Get: 获取单条结果集
func (m *Mysql) Get() *Rows {
	m.sql = "SELECT "+m.fields+" FROM "+m.currTable
	if len(m.where) > 0 {
		m.sql += " WHERE "+m.where
	}
	stmt,err := m.prepare(m.sql)
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

// GetOne: 获取单条结果集
func (m *Mysql) GetOne() *Row {
	m.sql = "SELECT "+m.fields+" FROM "+m.currTable+" WHERE "+m.where
	stmt,err := m.prepare(m.sql)
	if err != nil {
		log.Panic(err.Error())
	}
	row := stmt.QueryRow()
	m.free()
	return row
}

// Add：添加单条记录
func (m *Mysql) Add (addData ...interface{}) (Result,error) {
	var insertValue []interface{}
	var fields []string
	for _,data := range addData {
		v := data.([]interface{})
		fields = append(fields,v[0].(string))
		insertValue = append(insertValue,v[1])
	}
	m.buildInsertSql(fields)
	defer m.free()

	return m.insert(insertValue...)
}

// Adds： 批量添加数据
func (m *Mysql) Adds (addField []string,addValues ...interface{}) (int64,error){
	m.buildInsertSql(addField)
	var lastInsertId int64 = 0
	var err error = nil
	for _,value := range addValues {
		result,err := m.insert(value.([]interface{})...)
		if err != nil {
			break
		}
		lastInsertId , err = result.LastInsertId()
	}

	m.free()

	return lastInsertId, err

}

// Update: 更新记录
func (m *Mysql) Update(updateData ...interface{})(Result,error) {
	var updateValue []interface{}
	var fields []string
	for _,data := range updateData {
		v := data.([]interface{})
		fields = append(fields,v[0].(string))
		updateValue = append(updateValue,v[1])
	}
	m.buildInsertSql(fields)
	defer m.free()

	return m.update(updateValue...)
}

// Count: 获取指定条件下记录数量
func (m *Mysql) Count() (int64,error){
	m.sql = "SELECT count(*) FROM " + m.currTable
	if len(m.where) > 0 {
		m.sql += " WHERE "+ m.where
	}
	stmt,err := m.prepare(m.sql)
	if err != nil {
		return 0,err
	}
	row := stmt.QueryRow()

	var count int64
	err = row.Scan([]interface{}{&count}...)

	m.free()
	return count,nil
}

func (m *Mysql) prepare(sql string) (*Stmt,error) {
	if stmt,ok := m.stmt[m.sql]; ok {
		return stmt,nil
	}
	stmt,err := m.connector.Prepare(m.sql)
	if err == nil {
		m.stmt[m.sql] = stmt
	}
	return stmt,err
}

func (m *Mysql) buildInsertSql(fields []string){
	var insertField string
	var insertPrepare string
	m.sql = "INSERT INTO "+m.currTable+"("
	for _,field := range fields {
		if len(insertField) != 0 {
			insertField += ","
			insertPrepare += ","
		}
		insertField += "`"+field+"`"
		insertPrepare += "?"
	}
	m.sql += insertField+") VALUES ("+insertPrepare+")"
}

func (m *Mysql) buildUpdateSql(fields []string){
	var update string
	m.sql = "UPDATE "+m.currTable+" SET "
	for _,field := range fields {
		if len(update) != 0 {
			update += ","
		}
		update += "`"+field+"`=?"
	}
	m.sql += update + m.where
}

func (m Mysql) insert (addValue ...interface{}) (Result,error){
	stmt,err := m.prepare(m.sql)
	if err != nil {
		m.Err = err
		return nil,err
	}

	result , err := stmt.Exec(addValue...)
	if err != nil {
		m.Err = err
		return nil,err
	}
	return result, nil
}

func (m Mysql) update (updateValue ...interface{}) (Result,error){
	stmt,err := m.prepare(m.sql)
	if err != nil {
		m.Err = err
		return nil,err
	}

	result , err := stmt.Exec(updateValue...)
	if err != nil {
		m.Err = err
		return nil,err
	}
	return result, nil
}

func (m *Mysql) free() bool {
	for _,stmt := range m.stmt {
		err := stmt.Close()
		if err != nil {
			break
		}
	}
	m.stmt = make(map[string]*Stmt)
	m.sql = ""
	m.currTable = ""
	m.where = ""
	m.fields = ""

	server := reflect.ValueOf(Db.dbServers[DriverName]).Interface().(*MysqlServer)

	server.SetHandle(m)

	return true

}
