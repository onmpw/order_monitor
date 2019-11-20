package db

import (
	"database/sql"
)



var (
	DataSourceName      = ""
	JdDataSourceName    = ""
	UcDataSourceName    = ""
	LocalDataSourceName = ""

	Db      = &BaseDb{}
	LocalDb *sql.DB
	JdDB    *sql.DB
)

type BaseDb struct {
	Connector     	*sql.DB
	GetConnection 	func(source string) *sql.DB
	dbContainer   	map[string]BaseDbContract
	dataSource		map[string]string
}

type BaseDbContract interface {
	Connection(connection string) *sql.DB
}

func (db *BaseDb) Init() error {

	// 加载db配置项
	loadConfig()

	db.dbContainer = make(map[string]BaseDbContract)
	db.dataSource = make(map[string]string)

	NewMysql().registerDb()

	err := db.SwitchServer("mysql")

	return err
}

func (db *BaseDb) SwitchServer(driverName string) error{
	if contract, ok := db.dbContainer[driverName]; ok {
		db.GetConnection = contract.Connection
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
	db.dataSource[connection] = source

	return source
}
