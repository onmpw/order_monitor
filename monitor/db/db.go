package db



var Db      = &BaseDb{}

type BaseDb struct {
	Connector     	BaseDbContract
	GetConnection 	func(source string) BaseDbContract
	dbServers   	map[string]BaseDbContract
	dataSource		map[string]string
}

type BaseDbContract interface {
	Connection(connection string) BaseDbContract
	CheckDriverName(connection string) bool
	Table(table string)		BaseDbContract
}

func (db *BaseDb) Init() error {

	// 加载db配置项
	err := loadConfig()
	if err != nil {
		return err
	}

	db.dbServers = make(map[string]BaseDbContract)
	db.dataSource = make(map[string]string)

	NewMysql().registerDb()

	err = db.SwitchServer("mysql")

	return err
}

func (db *BaseDb) SwitchServer(driverName string) error{
	if contract, ok := db.dbServers[driverName]; ok {
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

func (db *BaseDb) Connect(source string) BaseDbContract {
	db.Connector = db.GetConnection(source)
	return db.Connector
}
