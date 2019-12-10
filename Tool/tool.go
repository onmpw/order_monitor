package Tool

import (
	"fmt"
	"monitor/monitor"
	"monitor/monitor/model"
	"reflect"
)

type OrderInfo struct {
}

func (o *OrderInfo) TableName() string {
	return "order_info"
}

type OrderUnusual struct {
	Type 		string
	Response 	string
	Remarks		string
}

func (o *OrderUnusual) TableName() string {
	return "order_sync_unusual"
}

func Init() {
	monitor.SafeCompanyOrder = monitor.NewSafeMap()
}

func CheckSync(order *monitor.MyOrderInfo, oriChan <-chan monitor.Jdp, quit <-chan int) error {

	var unusualCount int64

	for {
		select {
		case orderInfo := <-oriChan:
			order.TotalCount++ // 记录订单总数
			var succeedOrder monitor.ResultOrder
			var failedOrder monitor.ResultOrder // 临时存储失败订单信息
			count := model.Read(new(OrderInfo)).Filter("company_id",orderInfo.CompanyId).Filter("number",orderInfo.Oid).Filter("source","!=","System").Count()

			monitor.SafeCompanyOrder.NewElement(orderInfo)
			if count == 0 {
				order.FailedCount++ //  记录失败数量

				// 记录失败订单的信息
				SetOrder(&failedOrder,orderInfo)

				modelObj := model.Read(new(OrderUnusual)).Filter("tid",orderInfo.Oid).Filter("is_delete","N")
				unusualCount = modelObj.Count()
				var orderUnusual *OrderUnusual
				_ = modelObj.GetOne(&orderUnusual)
				if unusualCount == 0 {
					failedOrder.Reason = " 订单同步未成功，没有进入异常表，具体原因有待查明！" //订单失败原因
				} else {
					failedOrder.Reason = " 订单同步异常，异常类型：" + orderUnusual.Type + " 原因：" + orderUnusual.Remarks // 同步订单异常
				}

				fmt.Printf("%s %c[31;40;1m%s%c[0m \n", orderInfo.Oid, 0x1B, "订单同步失败", 0x1B)
				order.FailedOrder = append(order.FailedOrder, failedOrder)

				monitor.SafeCompanyOrder.SetOrder(failedOrder,false)
			} else if count > 1 {
				order.FailedCount++ // 订单同步成功，但是重复了，所以标记为失败

				// 加入失败订单，提示出来
				SetOrder(&failedOrder,orderInfo)

				failedOrder.Reason = " 订单同步成功，但是重复进入系统表，需要删除其中的一条！" //订单失败原因

				fmt.Printf("%s %c[32;40;1m%s%c[0m \n", orderInfo.Oid, 0x1B, "订单同步成功", 0x1B)
				order.FailedOrder = append(order.FailedOrder, failedOrder)

				monitor.SafeCompanyOrder.SetOrder(failedOrder,false)

			} else {
				SetOrder(&succeedOrder,orderInfo)

				succeedOrder.Reason = "成功!"

				order.SucceedCount++ // 记录成功数量
				fmt.Printf("%s %c[32;40;1m%s%c[0m \n", orderInfo.Oid, 0x1B, "订单同步成功", 0x1B)
				monitor.SafeCompanyOrder.SetOrder(succeedOrder,true)

			}
		case <-quit:
			monitor.C <- 1
			return nil
		}
	}
}

func SetOrder(order interface{},oldOrder interface{}){
	oldOrderValue := reflect.ValueOf(oldOrder)
	oldOrderInd := reflect.Indirect(oldOrderValue)

	nov := reflect.ValueOf(order)
	newOrderInd := reflect.Indirect(nov)

	setFieldValue(&newOrderInd,oldOrderInd,"Id","Id")
	setFieldValue(&newOrderInd,oldOrderInd,"Name","Name")
	setFieldValue(&newOrderInd,oldOrderInd,"Oid","Oid")
	setFieldValue(&newOrderInd,oldOrderInd,"Response","Response")
	setFieldValue(&newOrderInd,oldOrderInd,"CompanyId","Cid")
	setFieldValue(&newOrderInd,oldOrderInd,"CompanyId","CompanyId")
	setFieldValue(&newOrderInd,oldOrderInd,"Created","Created")
	setFieldValue(&newOrderInd,oldOrderInd,"Modified","Modified")
	setFieldValue(&newOrderInd,oldOrderInd,"OrderType","Type")
	setFieldValue(&newOrderInd,oldOrderInd,"OrderType","OrderType")
	setFieldValue(&newOrderInd,oldOrderInd,"ShopId","Sid")
	setFieldValue(&newOrderInd,oldOrderInd,"ShopId","ShopId")
}

func setFieldValue(order *reflect.Value,oldOrder reflect.Value,fieldName string,name string) {
	field := order.FieldByName(fieldName)

	if field.IsValid() && field.CanSet() {
		val := oldOrder.FieldByName(name)
		if val.IsValid() {
			field.Set(val)
		}
	}
}

func Close(){

}
