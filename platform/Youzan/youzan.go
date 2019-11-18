package Youzan

import (
	"database/sql"
	"log"
	"monitor/Tool"
	"monitor/platform"
)

var yzChan      = make(chan int, 1)

var Order = platform.MyOrderInfo{
	Platform: "有赞", PlatformKey:"YZ",
}

/**
 * 获取有赞的原始数据
 */
func getYouZanOriginData() (<-chan platform.Jdp, error) {
	var myT platform.MyTime
	youzanDb, err := sql.Open(platform.DriverName, platform.DataSourceName)

	if err != nil {
		return nil, err
	}

	defer platform.CloseDb(youzanDb)

	err = youzanDb.Ping()
	if err != nil {
		return nil, err
	}

	stmtOut, err := youzanDb.Prepare("select id,oid,response,cid,created,modified,`type`,sid from jdp_youzan_order_trade where modified>=? and modified<=?")
	if err != nil {
		return nil, err
	}
	defer platform.CloseStmt(stmtOut)

	// 计算时间
	myT.CalculateTime()

	var youzanJdp platform.Jdp

	var inter = platform.RowData{&youzanJdp.Id, &youzanJdp.Oid, &youzanJdp.Response, &youzanJdp.CompanyId, &youzanJdp.Created, &youzanJdp.Modified, &youzanJdp.OrderType, &youzanJdp.ShopId}

	rows, err := stmtOut.Query(myT.Start, myT.End)
	if err != nil {
		return nil, err
	}
	oriChannel := make(chan platform.Jdp)

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
