package Jd

import (
	"database/sql"
	"fmt"
	"log"
	"monitor/Tool"
	"monitor/platform"
)

var jdChan = make(chan int, 1)
var Order = platform.MyOrderInfo{
	Platform: "京东",
}

/**
 * 获取京东的原始数据
 */
func getJdOriginData() (<-chan platform.Jdp, error) {
	var myT platform.MyTime
	fmt.Println(platform.JdDataSourceName)
	jdDb, err := sql.Open(platform.DriverName, platform.JdDataSourceName)

	if err != nil {
		return nil, err
	}

	defer platform.CloseDb(jdDb)

	err = jdDb.Ping()
	if err != nil {
		return nil, err
	}

	stmtOut, err := jdDb.Prepare("select id,oid,response,cid,created,modified,`type`,sid from jdp_jd_order_trade where modified>=? and modified<=?")
	if err != nil {
		return nil, err
	}
	defer platform.CloseStmt(stmtOut)

	// 计算时间
	myT.CalculateTime()

	var jdJdp platform.Jdp

	var inter = platform.RowData{&jdJdp.Id, &jdJdp.Oid, &jdJdp.Response, &jdJdp.CompanyId, &jdJdp.Created, &jdJdp.Modified, &jdJdp.OrderType, &jdJdp.ShopId}

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
