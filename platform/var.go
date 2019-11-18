package platform

import "database/sql"

var (
	DriverName        = "mysql"
	DataSourceName    = ""
	JdDataSourceName  = ""
	UcDataSourceName  = ""
	LocalDataSourceName  = ""

	DateFormat  = "2006-01-02 15:04:05"
	TypeNum           = 7

	C = make(chan int, TypeNum) // channel 用于控制多协程

	CompanyMap = make(mis)
	ShopMap    = make(mis)

	//companyOrder = make(map[int]*myOrderInfo)
	SafeCompanyOrder *safeMap

	Db *sql.DB
	LocalDb *sql.DB

	CountStmt *sql.Stmt
	UnusualCountStmt *sql.Stmt

	InsertStmt *sql.Stmt

	Config         = make(map[string]string)
	ConfigFilePath = ""
)
