package Tool

import (
	"database/sql"
	"fmt"
	"log"
	"monitor/platform"
	"os"
)

func CheckSync(order *platform.MyOrderInfo, oriChan <-chan platform.Jdp, quit <-chan int) error {
	err := platform.Db.Ping()
	if err != nil {
		return err
	}

	order.FailedOrder = make([]platform.BadOrder, 0)
	/*stmtOut, err := platform.Db.Prepare("select count(*) from order_info where company_id=? and number=? and source!=\"System\"")
	stmUnusualOut, err := platform.Db.Prepare("select count(*),`type`,`response`,`remarks` from order_sync_unusual where tid=? and is_delete='N'")
	if err != nil {
		return err
	}
	defer platform.CloseStmt(stmtOut)
	defer platform.CloseStmt(stmUnusualOut)*/

	var count, unusualCount int
	var unusualType, response, remarks string

	for {
		select {
		case orderInfo := <-oriChan:
			order.TotalCount++ // 记录订单总数

			var failedOrder platform.BadOrder // 临时存储失败订单信息
			row := platform.CountStmt.QueryRow(orderInfo.CompanyId, orderInfo.Oid)
			err = row.Scan(&count)
			if err != nil {
				panic(err.Error()) // proper error handling instead of panic in your app
			}

			platform.SafeCompanyOrder.Lock()
			if _, ok := platform.SafeCompanyOrder.CompanyOrder[orderInfo.ShopId]; !ok {
				platform.SafeCompanyOrder.CompanyOrder[orderInfo.ShopId] = &platform.MyOrderInfo{}
			}
			platform.SafeCompanyOrder.CompanyOrder[orderInfo.ShopId].TotalCount++
			platform.SafeCompanyOrder.Unlock()
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

				unusualRow := platform.UnusualCountStmt.QueryRow(orderInfo.Oid)
				_ = unusualRow.Scan(&unusualCount, &unusualType, &response, &remarks)
				if unusualCount == 0 {
					failedOrder.Reason = " 订单同步未成功，没有进入异常表，具体原因有待查明！" //订单失败原因
				} else {
					failedOrder.Reason = " 订单同步异常，异常类型：" + unusualType + " 原因：" + remarks // 同步订单异常
				}

				res , _ := platform.InsertStmt.Exec("hpf","15344532121",5053,68381,"company_name","shop_name","WS",orderInfo.Oid,1001,"F",0.13,failedOrder.Reason)

				fmt.Printf("%s %c[31;40;1m%s%c[0m \n", orderInfo.Oid, 0x1B, "订单同步失败", 0x1B)
				fmt.Println(res)
				order.FailedOrder = append(order.FailedOrder, failedOrder)
				platform.SafeCompanyOrder.Lock()
				platform.SafeCompanyOrder.CompanyOrder[orderInfo.ShopId].FailedCount++
				platform.SafeCompanyOrder.CompanyOrder[orderInfo.ShopId].FailedOrder = append(platform.SafeCompanyOrder.CompanyOrder[orderInfo.ShopId].FailedOrder, failedOrder)
				platform.SafeCompanyOrder.Unlock()
			} else if count > 1 {
				order.FailedCount++ // 订单同步成功，但是重复了，所以标记为失败
				/*platform.SafeCompanyOrder.Lock()
				platform.SafeCompanyOrder.CompanyOrder[orderInfo.ShopId].SucceedCount++
				platform.SafeCompanyOrder.Unlock()*/

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
				platform.SafeCompanyOrder.Lock()
				platform.SafeCompanyOrder.CompanyOrder[orderInfo.ShopId].FailedCount++
				platform.SafeCompanyOrder.CompanyOrder[orderInfo.ShopId].FailedOrder = append(platform.SafeCompanyOrder.CompanyOrder[orderInfo.ShopId].FailedOrder, failedOrder)
				platform.SafeCompanyOrder.Unlock()

			} else {
				order.SucceedCount++ // 记录成功数量
				fmt.Printf("%s %c[32;40;1m%s%c[0m \n", orderInfo.Oid, 0x1B, "订单同步成功", 0x1B)
				platform.SafeCompanyOrder.Lock()
				platform.SafeCompanyOrder.CompanyOrder[orderInfo.ShopId].SucceedCount ++
				platform.SafeCompanyOrder.Unlock()

			}
		case <-quit:
			platform.C <- 1
			return nil
		}
	}
}


func ParseConfig() error{
	//platform.ConfigFilePath = "/etc/"
	platform.ConfigFilePath = "./"
	var configFileName = platform.ConfigFilePath+platform.ConfigFileName

	file,err := os.OpenFile(configFileName,os.O_RDONLY, 0755)
	if err != nil {
		log.Panic(err.Error())
	}

	length, err := file.Seek(0,2)
	if err != nil {
		log.Panic(err.Error())
	}
	var b = make([]byte,length)

	_, err = file.Seek(0,0)
	if err != nil {
		return err
	}

	_,err = file.Read(b)
	if err != nil {
		return err
	}
	// 读取成功之后，加入一个结束符
	b = append(b,byte('\n'))

	var byteStr = make([]byte,0)

	var commentFlag = false

	for _,char := range b {
		// # 作为注释
		if char == byte(platform.CommentIdentifier) {
			commentFlag = true // 表示此行是注释行
			continue
		}

		if char == byte('\n') {
			if commentFlag {
				// 表示刚刚读取的那一行是注释行，下面过程不用处理直接重置 byteStr
				byteStr = byteStr[0:0]
				commentFlag = false  // 重置标识符
				continue
			}

			if len(byteStr) == 0 {  // 空行 直接读取下一行
				continue
			}

			flag := 0
			keyEnd := 0
			valueStart := 0
			for index,t := range byteStr {
				if t == byte(' ') || t == byte('\t') {
					if flag == 0{
						keyEnd = index
					}
					flag = 1
				}
				if flag == 1 && !(t == byte(' ') || t == byte('\t')) {
					valueStart = index
					break
				}
			}
			platform.Config[string(byteStr[:keyEnd])] = string(byteStr[valueStart:])
			byteStr = byteStr[0:0]
			continue
		}
		if !commentFlag { // 注释行的字符不作处理
			byteStr = append(byteStr, char)
		}
	}
	return nil
}

func ParseDatabaseInfo() error{
	platform.DataSourceName = platform.Config["username"]+":"+platform.Config["password"]+"@tcp("+platform.Config["host"]+")/"+platform.Config["database"]
	platform.UcDataSourceName = platform.Config["username"]+":"+platform.Config["password"]+"@tcp("+platform.Config["host"]+")/"+platform.Config["uc_database"]
	platform.JdDataSourceName = platform.Config["jd_username"]+":"+platform.Config["jd_password"]+"@tcp("+platform.Config["jd_host"]+":"+platform.Config["jd_port"]+")/"+platform.Config["database"]
	platform.LocalDataSourceName = platform.Config["local_username"]+":"+platform.Config["local_password"]+"@tcp("+platform.Config["local_host"]+")/"+platform.Config["local_database"]

	platform.Db , _ = sql.Open(platform.DriverName, platform.DataSourceName)

	var err error
	platform.CountStmt, err = platform.Db.Prepare("select count(*) from order_info where company_id=? and number=? and source!=\"System\"")
	if err != nil {
		log.Panic(err.Error())
	}
	platform.UnusualCountStmt, err = platform.Db.Prepare("select count(*),`type`,`response`,`remarks` from order_sync_unusual where tid=? and is_delete='N'")
	if err != nil {
		log.Panic(err.Error())
	}

	platform.LocalDb , _ = sql.Open(platform.DriverName,platform.LocalDataSourceName)
	platform.InsertStmt , err = platform.LocalDb.Prepare("insert into `mihuan_order_monitor`(`receiver_name`,`receiver_mobile`,`company_id`,`shop_id`,`company_name`,`shop_name`,`order_type`,`order_id`,`platform_id`,`sync_result`,`price`,`description`) values (?,?,?,?,?,?,?,?,?,?,?,?)")

	if err != nil {
		log.Panic(err.Error())
	}

	return nil
}

func Config(key string) string {
	if value, ok := platform.Config[key]; ok {
		return value
	}
	return ""
}

func Close(){
	platform.CloseStmt(platform.CountStmt)
	platform.CloseStmt(platform.UnusualCountStmt)
	platform.CloseStmt(platform.InsertStmt)
	platform.CloseDb(platform.Db)
	platform.CloseDb(platform.LocalDb)
}
