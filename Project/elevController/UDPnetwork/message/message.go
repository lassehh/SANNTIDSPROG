package UDPnetwork
type Message struct {
	Target_floor int
	Current_floor int 
	Id int
	Timestamp int
	NewOrders[10] int
	Master bool
	MasterIP string
}
