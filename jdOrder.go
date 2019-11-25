package main

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"monitor/Tool"
	"monitor/monitor"
	"monitor/monitor/db"
	"monitor/platform/Alibb"
	"monitor/platform/Jd"
	"monitor/platform/Pdd"
	"monitor/platform/Wm"
	"monitor/platform/Youzan"
	"time"
)

var companyChan = make(chan int, 1)
var shopChan    = make(chan int, 1)

/**
 * 获取店铺信息
 */
func getShopInfo() (<-chan monitor.ShopInfo, error) {
	shopDb, err := sql.Open(monitor.DriverName, monitor.DataSourceName)
	if err != nil {
		return nil, err
	}

	defer monitor.CloseDb(shopDb)

	err = shopDb.Ping()
	if err != nil {
		return nil, err
	}

	stmtOut, err := shopDb.Prepare("select sid,name,alias,nick,`type` from shop_taobao where is_delete=0 and end_date > ?")
	if err != nil {
		return nil, err
	}

	defer monitor.CloseStmt(stmtOut)

	var shop monitor.ShopInfo
	var inter = monitor.RowData{&shop.ShopId, &shop.Name, &shop.Alias, &shop.Nick, &shop.ShopType}

	now := time.Unix(time.Now().Unix(), 0).Format(monitor.DateFormat)
	rows, err := stmtOut.Query(now)
	if err != nil {
		return nil, err
	}

	shopOri := make(chan monitor.ShopInfo)
	go func() {
		for rows.Next() {
			err = rows.Scan(inter...)
			if err != nil {
				panic(err.Error())
			}
			shopOri <- shop
		}
		shopChan <- 1
	}()

	return shopOri, nil
}

/**
 * 获取公司信息
 */
func getCompanyInfo() (<-chan monitor.CompanyInfo, error) {
	ucDb, err := sql.Open(monitor.DriverName, monitor.UcDataSourceName)
	if err != nil {
		return nil, err
	}

	defer monitor.CloseDb(ucDb)

	err = ucDb.Ping()
	if err != nil {
		return nil, err
	}

	stmtOut, err := ucDb.Prepare("select id,name from `company_detail` where `is_delete`=0")
	if err != nil {
		return nil, err
	}
	defer monitor.CloseStmt(stmtOut)

	var company monitor.CompanyInfo
	var inter = monitor.RowData{&company.Id, &company.Name}

	rows, err := stmtOut.Query()
	if err != nil {
		return nil, err
	}

	companyOri := make(chan monitor.CompanyInfo)
	go func() {
		for rows.Next() {
			err = rows.Scan(inter...)
			if err != nil {
				panic(err.Error())
			}
			companyOri <- company
		}
		companyChan <- 1
	}()
	return companyOri, nil

}

func ParseCompany() {
	oriChan, err := getCompanyInfo()
	if err != nil {
		log.Panic(err.Error())
	}
	go func() {
		for {
			select {
			case info := <-oriChan:
				monitor.CompanyMap[info.Id] = info.Name
			case <-companyChan:
				monitor.C <- 1
				return
			}
		}
	}()
}

func ParseShop() {
	oriChan, err := getShopInfo()
	if err != nil {
		log.Panic(err.Error())
	}
	go func() {
		for {
			select {
			case info := <-oriChan:

				monitor.ShopMap[info.ShopId] = info.Alias
			case <-shopChan:
				monitor.C <- 1
				return
			}
		}
	}()
}

type user struct {
	id 		int
	name 	string
	mobile 	string
}

type User struct{
	id		int
	name 	string
	mobile 	string
}

func main() {
	/*ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	select {
	case <-time.After(1 * time.Second):
		fmt.Println("overslept")
	case <-ctx.Done():
		fmt.Println(ctx.Err()) // prints "context deadline exceeded"
	}

	return*/
	err := monitor.Init()
	err = db.Db.Init()
	if err != nil {
		log.Panic(err.Error())
	}
	/*var user User
	rows := db.Db.Connector().Table("jiyi_user_info").Select("id","name","mobile").Where([]interface{}{[]interface{}{"id",17976}}...).GetOne()
	err = rows.Scan([]interface{}{&user.id,&user.name,&user.mobile}...)
	fmt.Println(user)*/
	/*for rows.Next() {
		err = rows.Scan([]interface{}{&user.id,&user.name,&user.mobile}...)
		fmt.Println(user.id)
	}*/
	//lastInsertId , err := db.Db.Connector().Table("order_info").Adds([]string{"oid","username"},[]interface{}{[]interface{}{3423312,"jiyi10"},[]interface{}{3423313,"jiyi11"}}...)
	var finishChan = make(chan int,1)
	var num = 12
	var finishedNum = 0
	for i := 0; i< num; i++ {
		go func() {
			lastInsertId , _ := db.Db.Connector().Table("order_info").Adds([]string{"oid","username"},[]interface{}{[]interface{}{3423316+i,"jiyia"+string(i)},[]interface{}{3423316+i+1,"jiyib"+string(i)}}...)
			finishChan <- 1
			fmt.Println(lastInsertId)
		}()
	}
	for {
		select {
		case <- finishChan :
			finishedNum++
			if finishedNum == num {
				return
			}
		}
	}

	return

	// 解析数据库连接信息
	_ = Tool.ParseDatabaseInfo()


	monitor.SafeCompanyOrder = monitor.NewSafeMap()

	defer close(monitor.C)
	//for num:=0 ; num < 10000; num++ {
	start := time.Unix(time.Now().Unix(), 0).Format(monitor.DateFormat)
	fmt.Println(start)

	go Jd.ParseJd()
	go Pdd.ParsePdd()
	go Wm.ParseWm()
	go Youzan.ParseYouZan()
	go Alibb.ParseAlibb()
	go ParseCompany()
	go ParseShop()

	monitor.Wait(monitor.C, monitor.TypeNum)

	end := time.Unix(time.Now().Unix(), 0).Format(monitor.DateFormat)
	fmt.Println(end)

	var showOrder monitor.Order = monitor.MyOrderInfoArr{Jd.Order, Pdd.Order, Wm.Order, Youzan.Order,Alibb.Order}

	showOrder.ShowOrderInfo()

	// 统计每个店铺同步订单失败情况
	for sid, order := range monitor.SafeCompanyOrder.CompanyOrder {
		fmt.Println(sid, "->", monitor.ShopMap[sid], ": 订单总共：", order.TotalCount, " 成功：", order.SucceedCount, " 失败：", order.FailedCount)
		for _, val := range order.FailedOrder {
			fmt.Println(val.Oid)
		}
	}

	//}
	Tool.Close()

}
