package db



var Db      = &BaseDb{}
const DefaultConnection = "production"   // 默认的连接名称  需要在配置文件(config.go)中添加其相应的配置项

type BaseDb struct {
	Connector     	BaseDbContract
	GetConnection 	func(source string) BaseDbContract
	dbServers   	map[string]BaseDbServer
	dataSource		map[string]string
}

type BaseDbServer interface {
	Connection(connection string) BaseDbContract
	CheckDriverName(connection string) bool
}

type BaseDbContract interface {
	Table(table string)		BaseDbContract
	GetTable()				string
	Select()				BaseDbContract
}

func (this *BaseDb) Init() error {

	// 加载db配置项
	err := loadConfig()
	if err != nil {
		return err
	}

	this.dbServers = make(map[string]BaseDbServer)
	this.dataSource = make(map[string]string)

	NewMysqlServer().registerDb()

	err = this.SwitchServer("mysql")

	return err
}

func (this *BaseDb) SwitchServer(driverName string) error{
	if server, ok := this.dbServers[driverName]; ok {
		this.Connector = server.Connection(DefaultConnection)
		this.GetConnection = server.Connection
		return nil
	}
	return &NoDriverError{errMsg:driverName+" driver not found!",Err:nil}
}

func (this *BaseDb) getDataSource(connection string) string {
	if dataSource , ok := this.dataSource[connection]; ok {
		return dataSource
	}
	localConn := connections[connection]
	var port = ""
	if val, ok := localConn["port"]; ok {
		port = ":"+val
	}
	source := localConn["username"]+":"+localConn["password"]+"@tcp("+localConn["host"]+port+")/"+localConn["database"]
	this.dataSource[connection] = source

	return source
}
