package Tool

import (
	"database/sql"
	"fmt"
	"log"
	"monitor/monitor"
	"monitor/monitor/config"
)

func CheckSync(order *monitor.MyOrderInfo, oriChan <-chan monitor.Jdp, quit <-chan int) error {
	err := monitor.Db.Ping()
	if err != nil {
		return err
	}

	order.FailedOrder = make([]monitor.BadOrder, 0)
	/*stmtOut, err := monitor.Db.Prepare("select count(*) from order_info where company_id=? and number=? and source!=\"System\"")
	stmUnusualOut, err := monitor.Db.Prepare("select count(*),`type`,`response`,`remarks` from order_sync_unusual where tid=? and is_delete='N'")
	if err != nil {
		return err
	}
	defer monitor.CloseStmt(stmtOut)
	defer monitor.CloseStmt(stmUnusualOut)*/

	var count, unusualCount int
	var unusualType, response, remarks string

	for {
		select {
		case orderInfo := <-oriChan:
			order.TotalCount++ // 记录订单总数

			var succeedOrder monitor.SucceedOrder
			var failedOrder monitor.BadOrder // 临时存储失败订单信息
			row := monitor.CountStmt.QueryRow(orderInfo.CompanyId, orderInfo.Oid)
			err = row.Scan(&count)
			if err != nil {
				panic(err.Error()) // proper error handling instead of panic in your app
			}

			monitor.SafeCompanyOrder.Lock()
			if _, ok := monitor.SafeCompanyOrder.CompanyOrder[orderInfo.ShopId]; !ok {
				monitor.SafeCompanyOrder.CompanyOrder[orderInfo.ShopId] = &monitor.MyOrderInfo{}
			}
			monitor.SafeCompanyOrder.CompanyOrder[orderInfo.ShopId].TotalCount++
			monitor.SafeCompanyOrder.Unlock()
			if count == 0 {
				order.FailedCount++ //  记录失败数量

				// 记录失败订单的信息
				failedOrder.Oid = orderInfo.Oid
				failedOrder.ShopId = orderInfo.ShopId
				failedOrder.CompanyId = orderInfo.CompanyId
				failedOrder.Created = orderInfo.Created
				failedOrder.Modified = orderInfo.Modified
				failedOrder.Response = orderInfo.Response
				failedOrder.OrderType = orderInfo.OrderType

				unusualRow := monitor.UnusualCountStmt.QueryRow(orderInfo.Oid)
				_ = unusualRow.Scan(&unusualCount, &unusualType, &response, &remarks)
				if unusualCount == 0 {
					failedOrder.Reason = " 订单同步未成功，没有进入异常表，具体原因有待查明！" //订单失败原因
				} else {
					failedOrder.Reason = " 订单同步异常，异常类型：" + unusualType + " 原因：" + remarks // 同步订单异常
				}

				fmt.Printf("%s %c[31;40;1m%s%c[0m \n", orderInfo.Oid, 0x1B, "订单同步失败", 0x1B)
				order.FailedOrder = append(order.FailedOrder, failedOrder)
				monitor.SafeCompanyOrder.Lock()
				monitor.SafeCompanyOrder.CompanyOrder[orderInfo.ShopId].FailedCount++
				monitor.SafeCompanyOrder.CompanyOrder[orderInfo.ShopId].FailedOrder = append(monitor.SafeCompanyOrder.CompanyOrder[orderInfo.ShopId].FailedOrder, failedOrder)
				monitor.SafeCompanyOrder.Unlock()
			} else if count > 1 {
				order.FailedCount++ // 订单同步成功，但是重复了，所以标记为失败
				/*monitor.SafeCompanyOrder.Lock()
				monitor.SafeCompanyOrder.CompanyOrder[orderInfo.ShopId].SucceedCount++
				monitor.SafeCompanyOrder.Unlock()*/

				// 加入失败订单，提示出来
				failedOrder.Oid = orderInfo.Oid
				failedOrder.ShopId = orderInfo.ShopId
				failedOrder.CompanyId = orderInfo.CompanyId
				failedOrder.Created = orderInfo.Created
				failedOrder.Modified = orderInfo.Modified
				failedOrder.Response = orderInfo.Response
				failedOrder.OrderType = orderInfo.OrderType

				failedOrder.Reason = " 订单同步成功，但是重复进入系统表，需要删除其中的一条！" //订单失败原因

				fmt.Printf("%s %c[32;40;1m%s%c[0m \n", orderInfo.Oid, 0x1B, "订单同步成功", 0x1B)
				order.FailedOrder = append(order.FailedOrder, failedOrder)
				monitor.SafeCompanyOrder.Lock()
				monitor.SafeCompanyOrder.CompanyOrder[orderInfo.ShopId].FailedCount++
				monitor.SafeCompanyOrder.CompanyOrder[orderInfo.ShopId].FailedOrder = append(monitor.SafeCompanyOrder.CompanyOrder[orderInfo.ShopId].FailedOrder, failedOrder)
				monitor.SafeCompanyOrder.Unlock()

			} else {
				succeedOrder.Oid = orderInfo.Oid
				succeedOrder.ShopId = orderInfo.ShopId
				succeedOrder.CompanyId = orderInfo.CompanyId
				succeedOrder.Created = orderInfo.Created
				succeedOrder.Modified = orderInfo.Modified
				succeedOrder.Response = orderInfo.Response
				succeedOrder.OrderType = orderInfo.OrderType
				succeedOrder.Reason = "成功!"

				order.SucceedCount++ // 记录成功数量
				fmt.Printf("%s %c[32;40;1m%s%c[0m \n", orderInfo.Oid, 0x1B, "订单同步成功", 0x1B)
				monitor.SafeCompanyOrder.Lock()
				monitor.SafeCompanyOrder.CompanyOrder[orderInfo.ShopId].SucceedCount ++
				monitor.SafeCompanyOrder.CompanyOrder[orderInfo.ShopId].SucceedOrder = append(monitor.SafeCompanyOrder.CompanyOrder[orderInfo.ShopId].SucceedOrder, succeedOrder)
				monitor.SafeCompanyOrder.Unlock()

			}
		case <-quit:
			monitor.C <- 1
			return nil
		}
	}
}


func ParseDatabaseInfo() error{
	monitor.DataSourceName = config.Conf.Config["username"]+":"+config.Conf.Config["password"]+"@tcp("+config.Conf.Config["host"]+")/"+config.Conf.Config["database"]
	monitor.UcDataSourceName = config.Conf.Config["username"]+":"+config.Conf.Config["password"]+"@tcp("+config.Conf.Config["host"]+")/"+config.Conf.Config["uc_database"]
	monitor.JdDataSourceName = config.Conf.Config["jd_username"]+":"+config.Conf.Config["jd_password"]+"@tcp("+config.Conf.Config["jd_host"]+":"+config.Conf.Config["jd_port"]+")/"+config.Conf.Config["database"]
	monitor.LocalDataSourceName = config.Conf.Config["local_username"]+":"+config.Conf.Config["local_password"]+"@tcp("+config.Conf.Config["local_host"]+")/"+config.Conf.Config["local_database"]

	monitor.Db , _ = sql.Open(monitor.DriverName, monitor.DataSourceName)

	var err error
	monitor.CountStmt, err = monitor.Db.Prepare("select count(*) from order_info where company_id=? and number=? and source!=\"System\"")
	if err != nil {
		log.Panic(err.Error())
	}
	monitor.UnusualCountStmt, err = monitor.Db.Prepare("select count(*),`type`,`response`,`remarks` from order_sync_unusual where tid=? and is_delete='N'")
	if err != nil {
		log.Panic(err.Error())
	}

	monitor.LocalDb , _ = sql.Open(monitor.DriverName,monitor.LocalDataSourceName)
	monitor.InsertStmt , err = monitor.LocalDb.Prepare("insert into `mihuan_order_monitor`(`receiver_name`,`receiver_mobile`,`company_id`,`shop_id`,`company_name`,`shop_name`,`order_type`,`order_id`,`platform_id`,`sync_result`,`price`,`description`) values (?,?,?,?,?,?,?,?,?,?,?,?)")

	if err != nil {
		log.Panic(err.Error())
	}

	return nil
}

func Config(key string) string {
	if value, ok := config.Conf.Config[key]; ok {
		return value
	}
	return ""
}

func Close(){
	monitor.CloseStmt(monitor.CountStmt)
	monitor.CloseStmt(monitor.UnusualCountStmt)
	monitor.CloseStmt(monitor.InsertStmt)
	monitor.CloseDb(monitor.Db)
	monitor.CloseDb(monitor.LocalDb)
}
