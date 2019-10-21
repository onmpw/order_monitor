package Wm

import (
	"database/sql"
	"log"
	"monitor/Tool"
	"monitor/platform"
)

var wmChan  = make(chan int, 1)
var Order = platform.MyOrderInfo{
	Platform: "微盟",
}

/**
 * 获取微盟的原始数据
 */
func getWmOriginData() (<-chan platform.Jdp, error) {
	var myT platform.MyTime
	wmDb, err := sql.Open(platform.DriverName, platform.DataSourceName)

	if err != nil {
		return nil, err
	}

	defer platform.CloseDb(wmDb)

	err = wmDb.Ping()
	if err != nil {
		return nil, err
	}

	stmtOut, err := wmDb.Prepare("select id,oid,response,cid,created,modified,`type`,sid from jdp_weimob_order_trade where modified>=? and modified<=?")
	if err != nil {
		return nil, err
	}
	defer platform.CloseStmt(stmtOut)

	// 计算时间
	myT.CalculateTime()

	var wmJdp platform.Jdp

	var inter = platform.RowData{&wmJdp.Id, &wmJdp.Oid, &wmJdp.Response, &wmJdp.CompanyId, &wmJdp.Created, &wmJdp.Modified, &wmJdp.OrderType, &wmJdp.ShopId}

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
