package Jd

import (
	"log"
	"monitor/Tool"
	"monitor/monitor"
	"monitor/monitor/db"
)

var jdChan = make(chan int, 1)
var Order = monitor.MyOrderInfo{
	Platform: "京东", PlatformKey:"JD",
}

/**
 * 获取京东的原始数据
 */
func getJdOriginData() (<-chan monitor.Jdp, error) {
	var myT monitor.MyTime
	// 计算时间
	myT.CalculateTime()
	var err error

	where := []interface{}{
		[]interface{}{"modified",">=",myT.Start},
		[]interface{}{"modified","<=",myT.End},
	}
	fields := []interface{}{
		"id","oid","response","cid","created","modified","type","sid",
	}
	rows := db.Db.GetConnection("jd_production").Table("jdp_jd_order_trade").Select(fields...).Where(where...).Get()
	var jdJdp monitor.Jdp

	var inter = monitor.RowData{&jdJdp.Id, &jdJdp.Oid, &jdJdp.Response, &jdJdp.CompanyId, &jdJdp.Created, &jdJdp.Modified, &jdJdp.OrderType, &jdJdp.ShopId}

	oriChannel := make(chan monitor.Jdp)

	go func() {
		for rows.Next() {
			err = rows.Scan(inter...)
			if err != nil {
				panic(err.Error())
			}
			oriChannel <- jdJdp
		}
		jdChan <- 1
	}()

	return oriChannel, nil
}

func ParseJd() {
	oriChan, err := getJdOriginData()
	if err != nil {
		log.Panic(err.Error())
	}
	go func() {
		err := Tool.CheckSync(&Order, oriChan, jdChan)
		if err != nil {
			log.Panic(err.Error())
		}
	}()
}
