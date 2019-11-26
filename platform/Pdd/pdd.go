package Pdd

import (
	"log"
	"monitor/Tool"
	"monitor/monitor"
	"monitor/monitor/db"
)

var pddChan = make(chan int, 1)
var Order = monitor.MyOrderInfo{
	Platform: "拼多多", PlatformKey:"PDD",
}

/**
 * 获取拼多多的原始数据
 */
func getPddOriginData() (<-chan monitor.Jdp, error) {
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
	rows := db.Db.Connector().Table("jdp_pdd_order_trade").Select(fields...).Where(where...).Get()
	var pddJdp monitor.Jdp

	var inter = monitor.RowData{&pddJdp.Id, &pddJdp.Oid, &pddJdp.Response, &pddJdp.CompanyId, &pddJdp.Created, &pddJdp.Modified, &pddJdp.OrderType, &pddJdp.ShopId}

	oriChannel := make(chan monitor.Jdp)

	go func() {
		for rows.Next() {
			err = rows.Scan(inter...)
			if err != nil {
				panic(err.Error())
			}
			oriChannel <- pddJdp
		}
		pddChan <- 1
	}()
	return oriChannel, nil
}

func ParsePdd() {
	oriChan, err := getPddOriginData()
	if err != nil {
		log.Panic(err.Error())
	}
	go func() {
		err := Tool.CheckSync(&Order, oriChan, pddChan)
		if err != nil {
			log.Panic(err.Error())
		}
	}()
}
