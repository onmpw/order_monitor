package Alibb

type OrderTrade struct {
	Id				int
	Oid				string
	Response		string
	Cid				int
	Created			string
	Modified		string
	Type 			string
	Sid 			int
}
//type OrderTrade monitor.Jdp


func (o *OrderTrade) TableName() string {
	return "jdp_alibb_order_trade"
}
