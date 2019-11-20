package Youzan

import (
	"database/sql"
	"log"
	"monitor/Tool"
	"monitor/monitor"
)

var yzChan      = make(chan int, 1)

var Order = monitor.MyOrderInfo{
	Platform: "有赞", PlatformKey:"YZ",
}

/**
 * 获取有赞的原始数据
 */
func getYouZanOriginData() (<-chan monitor.Jdp, error) {
	var myT monitor.MyTime
	youzanDb, err := sql.Open(monitor.DriverName, monitor.DataSourceName)

	if err != nil {
		return nil, err
	}

	defer monitor.CloseDb(youzanDb)

	err = youzanDb.Ping()
	if err != nil {
		return nil, err
	}

	stmtOut, err := youzanDb.Prepare("select id,oid,response,cid,created,modified,`type`,sid from jdp_youzan_order_trade where modified>=? and modified<=?")
	if err != nil {
		return nil, err
	}
	defer monitor.CloseStmt(stmtOut)

	// 计算时间
	myT.CalculateTime()

	var youzanJdp monitor.Jdp

	var inter = monitor.RowData{&youzanJdp.Id, &youzanJdp.Oid, &youzanJdp.Response, &youzanJdp.CompanyId, &youzanJdp.Created, &youzanJdp.Modified, &youzanJdp.OrderType, &youzanJdp.ShopId}

	rows, err := stmtOut.Query(myT.Start, myT.End)
	if err != nil {
		return nil, err
	}
	oriChannel := make(chan monitor.Jdp)

	go func() {
		for rows.Next() {
			err = rows.Scan(inter...)
			if err != nil {
				panic(err.Error())
			}
			oriChannel <- youzanJdp
		}
		yzChan <- 1
	}()
	return oriChannel, nil
}

func ParseYouZan() {
	oriCha, err := getYouZanOriginData()
	if err != nil {
		log.Panic(err.Error())
	}
	go func() {
		err := Tool.CheckSync(&Order, oriCha, yzChan)
		if err != nil {
			log.Panic(err.Error())
		}
	}()
}
