package main

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"monitor/Tool"
	"monitor/platform"
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
func getShopInfo() (<-chan platform.ShopInfo, error) {
	shopDb, err := sql.Open(platform.DriverName, platform.DataSourceName)
	if err != nil {
		return nil, err
	}

	defer platform.CloseDb(shopDb)

	err = shopDb.Ping()
	if err != nil {
		return nil, err
	}

	stmtOut, err := shopDb.Prepare("select sid,name,alias,nick,`type` from shop_taobao where is_delete=0 and end_date > ?")
	if err != nil {
		return nil, err
	}

	defer platform.CloseStmt(stmtOut)

	var shop platform.ShopInfo
	var inter = platform.RowData{&shop.ShopId, &shop.Name, &shop.Alias, &shop.Nick, &shop.ShopType}

	now := time.Unix(time.Now().Unix(), 0).Format(platform.DateFormat)
	rows, err := stmtOut.Query(now)
	if err != nil {
		return nil, err
	}

	shopOri := make(chan platform.ShopInfo)
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
func getCompanyInfo() (<-chan platform.CompanyInfo, error) {
	ucDb, err := sql.Open(platform.DriverName, platform.UcDataSourceName)
	if err != nil {
		return nil, err
	}

	defer platform.CloseDb(ucDb)

	err = ucDb.Ping()
	if err != nil {
		return nil, err
	}

	stmtOut, err := ucDb.Prepare("select id,name from `company_detail` where `is_delete`=0")
	if err != nil {
		return nil, err
	}
	defer platform.CloseStmt(stmtOut)

	var company platform.CompanyInfo
	var inter = platform.RowData{&company.Id, &company.Name}

	rows, err := stmtOut.Query()
	if err != nil {
		return nil, err
	}

	companyOri := make(chan platform.CompanyInfo)
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
				platform.CompanyMap[info.Id] = info.Name
			case <-companyChan:
				platform.C <- 1
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

				platform.ShopMap[info.ShopId] = info.Alias
			case <-shopChan:
				platform.C <- 1
				return
			}
		}
	}()
}

func main() {

	// 解析配置
	Tool.ParseConfig()

	// 解析数据库连接信息
	Tool.ParseDatabaseInfo()


	platform.SafeCompanyOrder = platform.NewSafeMap()

	defer platform.CloseDb(platform.Db)
	defer close(platform.C)
	//for num:=0 ; num < 10000; num++ {
	start := time.Unix(time.Now().Unix(), 0).Format(platform.DateFormat)
	fmt.Println(start)

	go Jd.ParseJd()
	go Pdd.ParsePdd()
	go Wm.ParseWm()
	go Youzan.ParseYouZan()
	go Alibb.ParseAlibb()
	go ParseCompany()
	go ParseShop()

	platform.Wait(platform.C, platform.TypeNum)

	end := time.Unix(time.Now().Unix(), 0).Format(platform.DateFormat)
	fmt.Println(end)

	var showOrder platform.Order = platform.MyOrderInfoArr{Jd.Order, Pdd.Order, Wm.Order, Youzan.Order,Alibb.Order}

	showOrder.ShowOrderInfo()

	// 统计每个店铺同步订单失败情况
	/*for sid, order := range safeCompanyOrder.companyOrder {
		fmt.Println(sid, "->", shopMap[sid], ": 订单总共：", order.totalCount, " 成功：", order.succeedCount, " 失败：", order.failedCount)
		for _, val := range order.failedOrder {
			fmt.Println(val.oid)
		}
	}*/
	//}

}
