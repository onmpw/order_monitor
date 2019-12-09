package Wm

import (
	"log"
	"monitor/Tool"
	"monitor/monitor"
	"monitor/monitor/model"
)

var wmChan  = make(chan int, 1)
var Order = monitor.MyOrderInfo{
	Platform: "微盟", PlatformKey:"WM",
}

/**
 * 获取微盟的原始数据
 */
func getWmOriginData() (<-chan monitor.Jdp, error) {
	var myT monitor.MyTime

	myT.CalculateTime()
	//var err error

	/*where := []interface{}{
		[]interface{}{"modified",">=",myT.Start},
		[]interface{}{"modified","<=",myT.End},
	}
	fields := []string{
		"id","oid","response","cid","created","modified","type","sid",
	}
	rows := db.Db.Connector().Table("jdp_weimob_order_trade").Select(fields...).Where(where...).Get()*/
	var trades []*OrderTrade

	num, _ := model.Read(new(OrderTrade)).Filter("modified",">=",myT.Start).Filter("modified","<=",myT.End).GetAll(&trades)

	var wmJdp monitor.Jdp

	//var inter = monitor.RowData{&wmJdp.Id, &wmJdp.Oid, &wmJdp.Response, &wmJdp.CompanyId, &wmJdp.Created, &wmJdp.Modified, &wmJdp.OrderType, &wmJdp.ShopId}

	oriChannel := make(chan monitor.Jdp)

	go func() {
		for i:=0; i < int(num); i++{
			Tool.SetOrder(&wmJdp,trades[i])

			oriChannel <- wmJdp
		}
		wmChan <- 1
	}()

	return oriChannel, nil
}

func ParseWm() {
	oriChan, err := getWmOriginData()
	if err != nil {
		log.Panic(err.Error())
	}
	go func() {
		err := Tool.CheckSync(&Order, oriChan, wmChan)
		if err != nil {
			log.Panic(err.Error())
		}
	}()
}
