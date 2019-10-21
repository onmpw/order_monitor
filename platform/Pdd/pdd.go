package Pdd

import (
	"database/sql"
	"log"
	"monitor/platform"
	"monitor/Tool"
)

var pddChan = make(chan int, 1)
var Order = platform.MyOrderInfo{
	Platform: "拼多多",
}

/**
 * 获取拼多多的原始数据
 */
func getPddOriginData() (<-chan platform.Jdp, error) {
	var myT platform.MyTime
	pddDb, err := sql.Open(platform.DriverName, platform.DataSourceName)

	if err != nil {
		return nil, err
	}

	defer platform.CloseDb(pddDb)

	err = pddDb.Ping()
	if err != nil {
		return nil, err
	}

	stmtOut, err := pddDb.Prepare("select id,oid,response,cid,created,modified,`type`,sid from jdp_pdd_order_trade where modified>=? and modified<=?")
	if err != nil {
		return nil, err
	}
	defer platform.CloseStmt(stmtOut)

	// 计算时间
	myT.CalculateTime()

	var pddJdp platform.Jdp

	var inter = platform.RowData{&pddJdp.Id, &pddJdp.Oid, &pddJdp.Response, &pddJdp.CompanyId, &pddJdp.Created, &pddJdp.Modified, &pddJdp.OrderType, &pddJdp.ShopId}

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
