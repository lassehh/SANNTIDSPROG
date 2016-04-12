//need to include message.go
package networkController
import(
"net"
."./UDPnetwork"
"strconv"
)
const PORT int = 40000;//setting the portnumber. Selected a random port

type Elevator_System struct {
	self_id int
	self_IP string
	elevators map[int]*Elevator //elevator declared in FSM
	//elevator_orders [10]int
	master int 
	MasterIP string
}

//initializing of elevator
func Initialize_elev() Elevator_System{
	var e Elevator_System
	e.elevators = make(map[int]*Elevator) 


	//initialize message here ?
	//initialize driver
	addr,_ :=net.InterfaceAddrs()
	tempVar:=addr[1]
 	ip:=tempVar.String()
	e.self_IP=ip[0:15]
	e.self_id := int(addr[1].String()[12]-'0')*100 + int(addr[1].String()[13]-'0')*10 + int(addr[1].String()[14]-'0')//this will work for IP-addresses of format ###.###.###.###, but not with only for ###.###.###.##/
	e.elevators[e.self_id]=new(Elevator)

	e.set_master()


}


func Is_elev_master() bool{
 	isMaster:= false
 	if e.self_id==e.master
 		{	
 			isMaster=true	
 		}
 	return isMaster
}


func Get_Master_IP()string{
	return e.MasterIP
}

/*func Initialize_connections(){
 	var isMaster bool =false
 	var tempIP string= e.self_IP[0:12]
 	var masterIP string ="255.255.255.255"
 	masterIP=tempIP + strconv.Itoa(e.master)
 	if e.self_id==e.master
 		{	
 			isMaster=true	
 		}
 	UDP_initialize(isMaster,PORT,masterIP)
	//don't think we're gonna need this one
	
}*/

func (e *Elevator_System) Set_floor(message Message) {
	e.elevators[message.Id].Floor = message.Current_floor_location
}

func MessageSetter(Broadcast_Message_chan chan Message,e *Elevator_System, elevat *Elevator ){
	
	var msg Message
	msg.Target_floor=elevat.DESTINATION_FLOOR
	msg.Current_floor=elevat.CURRENT_FLOOR
	//To Lasse: Need to agree on what the Type Elevator and elev needs to include
	msg.Id=e.self_id
	msg.MasterIP=MasterIP
	//Where to set Timestamp?
	msg.NewOrders=orders

	if (e.self_id==e.master){
		msg.Master=true
	}else{
		msg.Master=false
	}
	Broadcast_Message_chan<-msg
}



/* Functions needed to make:
	elev_remove
	evel_add
	broadcast_message
	queue_editor 	
*/

func queue_editor(){
	//want to prioritize orders from within the elevator

}	

func Message_Compiler(To_Master_Chan chan Message){
	var message_Reiceved=make([] Message)

	for {
		message_Received <- To_Master_Chan
	}
	

	//e :=elev_system[message_Received.Id]
	//e.INTERNAL_ORDERS=message_Received.NewOrders

}

func (e *Elevator_System) Remove_elev(id int) {
	delete(e.elevators, id)
	fmt.Println("Elevator ", id, " removed from network")

	e.select_master()

}

func elev_add(){
	//need to respond on some sort of button/message

}


func broadcast_message(msg chan Message){
	go UDP_send(PORT,msg)
}


func(e *Elevator_System) Set_master(){
	// checking which elevator has the highest IP to determine master 
	max :=0
	for i,_ :=range(e.elevators){
		if max<i  {
			max=i  
				
		}
	}
	e.master=max
	
	var tempIP string= e.self_IP[0:12]
 	e.MasterIP=tempIP + strconv.Itoa(e.master)
	fmt.Println("new master is", e.master)

}
