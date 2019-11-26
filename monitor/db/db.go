package db

import (
	. "database/sql"
	"sync"
)

var Db      = &BaseDb{}
const DefaultConnection = "production"   // 默认的连接名称  需要在配置文件(config.go)中添加其相应的配置项

type BaseDb struct {
	sync.RWMutex
	GetConnection 	func(source string) BaseDbContract
	dbServers   	map[string]BaseDbServer
	dataSource		map[string]string
}

type BaseDbServer interface {
	Connection(connection string) BaseDbContract
	CheckDriverName(connection string) bool
}

type BaseDbContract interface {
	Table(table string)									BaseDbContract
	GetTable()											string
	Select(fields ...interface{})						BaseDbContract
	Where(where ...interface{})							BaseDbContract
	Get()												*Rows
	GetOne()											*Row
	Add(addData ...interface{})							(Result,error)
	Adds (addField []string,addValues ...interface{}) 	(int64,error)
	Update(updateData ...interface{})					(Result,error)
	Count()												(int64,error)
}

func (db *BaseDb) Init() error {

	// 加载db配置项
	err := loadConfig()
	if err != nil {
		return err
	}

	db.dbServers = make(map[string]BaseDbServer)
	db.dataSource = make(map[string]string)

	NewMysqlServer().registerDb()

	err = db.SwitchServer("mysql")

	return err
}

func (db *BaseDb) SwitchServer(driverName string) error{
	if server, ok := db.dbServers[driverName]; ok {
		db.GetConnection = server.Connection
		return nil
	}
	return &NoDriverError{errMsg:driverName+" driver not found!",Err:nil}
}

func (db *BaseDb) getDataSource(connection string) string {
	if dataSource , ok := db.dataSource[connection]; ok {
		return dataSource
	}
	localConn := connections[connection]
	var port = ""
	if val, ok := localConn["port"]; ok {
		port = ":"+val
	}
	source := localConn["username"]+":"+localConn["password"]+"@tcp("+localConn["host"]+port+")/"+localConn["database"]
	db.Lock()
	db.dataSource[connection] = source
	db.Unlock()

	return source
}

func (db *BaseDb) Connector() BaseDbContract {
	Connect := db.GetConnection(DefaultConnection)
	return Connect
}