package main

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"monitor/Tool"
	"monitor/monitor"
	"monitor/monitor/db"
	"monitor/monitor/model"
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
	now := time.Unix(time.Now().Unix(), 0).Format(monitor.DateFormat)

	var shop monitor.ShopInfo
	var inter = monitor.RowData{&shop.ShopId, &shop.Name, &shop.Alias, &shop.Nick, &shop.ShopType}
	where := []interface{}{
		[]interface{}{"is_delete",0},
		[]interface{}{"end_date",">",now},
	}
	var err error
	rows := db.Db.Connector().Table("shop_taobao").Select("sid","name","alias","nick","type").Where(where...).Get()
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
	var company monitor.CompanyInfo
	var inter = monitor.RowData{&company.Id, &company.Name}

	var err error
	rows := db.Db.GetConnection("uc_production").Table("company_detail").Select("id","name").Where([]interface{}{[]interface{}{"is_delete",0}}...).Get()

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
type User struct {
	id 		int64
	name 	string
	mobile	string
}

func main() {
	model.RegisterModel(new(User))
	return
	err := monitor.Init()
	err = db.Db.Init()
	if err != nil {
		log.Panic(err.Error())
	}

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
