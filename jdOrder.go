package main

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/onmpw/JYGO/model"
	"log"
	"monitor/Tool"
	"monitor/monitor"
	"monitor/platform/Alibb"
	"monitor/platform/Jd"
	"monitor/platform/Pdd"
	"monitor/platform/Wm"
	"monitor/platform/Youzan"
	"time"
)

var companyChan = make(chan int, 1)
var shopChan    = make(chan int, 1)

type CompanyDetail struct {
	Id 			int
	Name 		string
}

func (c *CompanyDetail) TableName() string {
	return "company_detail"
}

func (c *CompanyDetail) Connection() string {
	return "uc_production"
}


type ShopInfo struct {
	Sid 		int
	Name 		string
	Alias 		string
	Nick		string
	Type 		string
}

func (s *ShopInfo) TableName() string {
	return "shop_taobao"
}

// 获取店铺信息
func getShopInfo() (<-chan ShopInfo, error) {
	now := time.Unix(time.Now().Unix(), 0).Format(monitor.DateFormat)

	shopOri := make(chan ShopInfo)

	var trades []*ShopInfo
	num , _ := model.Read(new(ShopInfo)).Filter("is_delete",0).Filter("end_date",">",now).GetAll(&trades)
	go func() {
		for i:=0; i < int(num); i++{
			shopOri<- *(trades[i]) //company
		}
		shopChan <- 1
	}()

	return shopOri, nil
}

// 获取公司信息
func getCompanyInfo() (<-chan CompanyDetail, error) {

	var trades []*CompanyDetail
	num,_ := model.Read(new(CompanyDetail)).Filter("is_delete",0).GetAll(&trades)
	companyOri := make(chan CompanyDetail)
	go func() {
		for i:=0; i < int(num); i++{
			companyOri <- *(trades[i]) //company
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

				monitor.ShopMap[info.Sid] = info.Alias
			case <-shopChan:
				monitor.C <- 1
				return
			}
		}
	}()
}

func main() {
	ModelInit()
	//err := monitor.Init()
	//err = db.Db.Init()
	/*if err != nil {
		log.Panic(err.Error())
	}*/

	Tool.Init()

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

func ModelInit() {
	model.Init()
	model.RegisterModel(
		new(Alibb.OrderTrade),
		new(Youzan.OrderTrade),
		new(Pdd.OrderTrade),
		new(Wm.OrderTrade),
		new(Jd.OrderTrade),
		new(Tool.OrderInfo),
		new(Tool.OrderUnusual),
		new(CompanyDetail),
		new(ShopInfo))
}
