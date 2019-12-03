package Alibb

import (
	"log"
	"monitor/Tool"
	"monitor/monitor"
	"monitor/monitor/db"
)


var alibbChan	= make(chan int, 1)
var Order = monitor.MyOrderInfo{
	Platform: "1688", PlatformKey:"1688",
}

/**
 * 获取1688的原始数据
 */
func getAlibbOriginData() (<-chan monitor.Jdp, error) {
	var myT monitor.MyTime

	myT.CalculateTime()
	var err error

	where := []interface{}{
		[]interface{}{"modified",">=",myT.Start},
		[]interface{}{"modified","<=",myT.End},
	}
	fields := []interface{}{
		"id","oid","response","cid","created","modified","type","sid",
	}
	rows := db.Db.Connector().Table("jdp_alibb_order_trade").Select(fields...).Where(where...).Get()

	var alibbJdp monitor.Jdp
	var inter = monitor.RowData{&alibbJdp.Id,&alibbJdp.Oid,&alibbJdp.Response,&alibbJdp.CompanyId,&alibbJdp.Created,&alibbJdp.Modified,&alibbJdp.OrderType,&alibbJdp.ShopId}

	oriChannel := make(chan monitor.Jdp)

	go func(){
		for rows.Next() {
			err = rows.Scan(inter...)
			if err != nil {
				panic(err.Error())
			}
			oriChannel <- alibbJdp
		}
		alibbChan <- 1
	}()
	return oriChannel,nil
}

func ParseAlibb() {
	oriCha, err := getAlibbOriginData()
	if err != nil {
		log.Panic(err.Error())
	}
	go func() {
		err := Tool.CheckSync(&Order,oriCha, alibbChan)
		if err != nil {
			log.Panic(err.Error())
		}
	}()
}
