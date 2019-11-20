package monitor

import (
	"database/sql"
	"fmt"
	"monitor/monitor/config"
	"sync"
	"time"
)

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
)

type Order interface {
	ShowOrderInfo()
}

type SucceedOrder struct {
	Oid, OrderType, Response, Created, Modified, Reason string
	CompanyId, ShopId                                   int
}

type BadOrder struct {
	Oid, OrderType, Response, Created, Modified, Reason string
	CompanyId, ShopId                                   int
}

type MyOrderInfo struct {
	TotalCount, FailedCount, SucceedCount int
	FailedOrder                           []BadOrder
	SucceedOrder                          []SucceedOrder
	Platform , PlatformKey                string
}

type MyOrderInfoArr []MyOrderInfo

func (myOrders MyOrderInfoArr) ShowOrderInfo() {
	for _, myOrder := range myOrders {
		fmt.Println(myOrder.Platform, "订单总共：", myOrder.TotalCount, " 成功：", myOrder.SucceedCount, " 失败：", myOrder.FailedCount)

		if myOrder.FailedCount > 0 {
			for index := 0; index < myOrder.FailedCount; index++ {
				fmt.Println(myOrder.FailedOrder[index].Oid, myOrder.FailedOrder[index].Reason, " 公司ID：", myOrder.FailedOrder[index].CompanyId, " 公司名称：", CompanyMap[myOrder.FailedOrder[index].CompanyId], " 店铺名称：", ShopMap[myOrder.FailedOrder[index].ShopId], " 店铺ID：", myOrder.FailedOrder[index].ShopId, " 订单类型：", myOrder.FailedOrder[index].OrderType)
			}
		}
	}
}

type CompanyInfo struct {
	Id   int
	Name string
}

type ShopInfo struct {
	ShopId                      int
	Name, Nick, Alias, ShopType string
}

// 原始数据中的字段
type Jdp struct {
	Id, CompanyId, ShopId                       int
	Oid, Response, Created, Modified, OrderType string
}

type MyTime struct {
	Start string
	End   string
}

type mis map[int]string

type RowData []interface{}

type safeMap struct {
	sync.RWMutex
	CompanyOrder map[int]*MyOrderInfo
}

func NewSafeMap() *safeMap {
	sm := new(safeMap)
	sm.CompanyOrder = make(map[int]*MyOrderInfo)
	return sm
}

func Init() error {

	// 加载配置文件
	err := config.Init()
	if err != nil {
		return err
	}
	return nil
}

/**
 * 等待goroutine 执行完
 */
func Wait(ch <-chan int, num int) {

	for i := 0; i < num; i++ {
		<-ch
	}
}

/**
 * 关闭数据库链接
 */
func CloseDb(db *sql.DB) {
	err := db.Close()
	if err != nil {
		panic(err.Error())
	}
}

/**
 * 关闭预处理context
 */
func CloseStmt(stmt *sql.Stmt) {
	err := stmt.Close()
	if err != nil {
		panic(err.Error())
	}
}

/**
 * 计算时间
 */
func (t *MyTime) CalculateTime() {
	now := time.Now()
	// 计算时间区间
	t.Start = time.Unix(now.Unix()-60*60*24, 0).Format(DateFormat)
	t.End = time.Unix(now.Unix(), 0).Format(DateFormat)
}
