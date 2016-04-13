//need to include message.go
package elevController
import(
	"net"
	"time"
	"encoding/json"
	"strconv"
	"os"
	"fmt"
)
const PORT int = 40000;//setting the portnumber. Selected a random port

type Elevator_System struct {
	selfID int
	selfIP string
	elevators map[int]*Elevator //elevator declared in FSM
	//elevator_orders [10]int
	master int 
	masterIP string
	timestamp int
}

type Message struct {
	destinationFloor int
	currentFloor int 
	ID int
	timestamp int
	InternalOrders [10]Button
	master bool
	masterIP string
}

func Initialize_elev_system() Elevator_System{
	var e_system Elevator_System
	e_system.elevators = 						make(map[int]*Elevator) 
	addr,_ := 									net.InterfaceAddrs()
	tempVar := 									addr[1]
 	ip := 										tempVar.String()
	e_system.selfIP = 							ip[0:15]
	e_system.selfID = 							int(addr[1].String()[12]-'0')*100 + int(addr[1].String()[13]-'0')*10 + int(addr[1].String()[14]-'0')//this will work for IP-addresses of format ###.###.###.###, but not with only for ###.###.###.##/
	e_system.elevators[e_system.selfID] = 		new(Elevator)
	Set_master(&e_system)
	return e_system
}

func Is_elev_master(e_system Elevator_System) bool{
 	isMaster:= false
 	if e_system.selfID==e_system.master{	
 			isMaster=true	
 		}
 	return isMaster
}

func Get_Master_IP(e_system Elevator_System)string{
	return e_system.masterIP
}

/*func Set_floor(message Message,e *Elevator_System) {
	e.elevators[message.Id].CURRENT_FlOOR = message.Current_floor
}*/

func MessageSetter(Broadcast_Message_chan chan Message, e_system *Elevator_System, e *Elevator ){
	
	var msg Message
	msg.destinationFloor = 		e.DestinationFloor
	msg.currentFloor = 			e.CurrentFloor
	msg.ID = 					e_system.selfID
	msg.masterIP = 				e_system.masterIP
	msg.InternalOrders = 		e_system.elevators[e_system.selfID].InternalOrders

	if (e_system.selfID == e_system.master){
		msg.master = true
	}else{
		msg.master = false
	}
	Broadcast_Message_chan<-msg
}

func Message_Compiler_Master(msgFromSlave Message, e_system *Elevator_System){
	if msgFromSlave.timestamp == e_system.timestamp {
		e_system.elevators[msgFromSlave.ID].InternalOrders = msgFromSlave.InternalOrders
	}
}

func Remove_elev(ID int, e *Elevator_System) {
	delete(e.elevators, ID)
	fmt.Println("Elevator ", ID, " removed from network")
	Set_master(e)
}

func Set_master(e_system *Elevator_System){
	// Checking which elevator has the highest IP to determine who is the master 
	max :=0
	for i,_ :=range(e_system.elevators){
		if max<i  {
			max=i  
				
		}
	}
	e_system.master = 			max
	e_system.timestamp = 		1
	var tempIP string = 		e_system.selfIP[0:12]

 	e_system.masterIP = 		tempIP + strconv.Itoa(e_system.master)
	fmt.Println("new master is", e_system.master)
}

func Timer(Timer_Chan chan bool){
	timer := time.NewTimer(time.Millisecond * 100)	
	<- timer.C
	Timer_Chan <- true
}

/* A Simple function to verify error */
func CheckError(err error) {
	if err != nil {
		fmt.Println("Error: ", err)
		os.Exit(0)
	}
}

func UDPListen(isMaster bool, listenPort int, masterIP string, rchvChan chan Message) {

	/* For testing: sett addresse lik ip#255:30000*/
	//ServerAddr, err := net.ResolveUDPAddr("udp", ":40000")
	ServerAddr, err := net.ResolveUDPAddr("udp", ":"+strconv.Itoa(listenPort))
	CheckError(err)

	/* Now listen at selected port */
	ServerConn, err := net.ListenUDP("udp", ServerAddr)
	CheckError(err)
	defer ServerConn.Close()
	

	var received_message Message 

	//var storageChannel := make(chan Message)

	buffer := make([]byte, 1024)
	trimmed_buffer:= make([]byte, 1)
		if (isMaster){
			for {
					n, addr, err := ServerConn.ReadFromUDP(buffer)
					trimmed_buffer = buffer[0:n]	
					fmt.Println("Received ", string(buffer[0:n]), " from ", addr)
					CheckError(err)

					err= json.Unmarshal(trimmed_buffer,&received_message)
					CheckError(err)

					//storageChannel <- received_message
					rchvChan<- received_message
					time.Sleep(time.Second * 1)
			}
		}else{
			for { // what if the elevators crashes and the slave becomes the master
				n, addr, err := ServerConn.ReadFromUDP(buffer)
				if (addr.String() == masterIP){
					trimmed_buffer = buffer[0:n]
					fmt.Println("Received ", string(buffer[0:n]), " from ", addr)
					CheckError(err)

					err= json.Unmarshal(trimmed_buffer,&received_message)
					CheckError(err)

					//storageChannel <- received_message
					rchvChan<-received_message
					time.Sleep(time.Second * 1)
				}
			
			}
		}
	
	}


 
//need to include message-sending
//if we want to send Elevator_System, we need to change the hierarchy
/*	
func UDPSend(transmitPort int,broadcastMessage chan Message,broadcastElevator_System chan broadcastElevator_System) {
	/* Dial up UDP */
	/*isMaster:= false
	for{
		select{
			case msg:= <- broadcastMessage:
				isMaster=false	//only slaves broadcast Messages
			case elev_system:=<-broadcastElevator_System:
				isMaster=true //only the master broadcast Elevator Systems
		}
		if (isMaster){
			BroadcastAddr, err := net.ResolveUDPAddr("udp", "255.255.255.255:"+strconv.Itoa(transmitPort))
			CheckError(err)

			/* Create a connection to the server */
			/*Conn, err := net.DialUDP("udp", nil, BroadcastAddr)
			CheckError(err)

			fmt.Println("This is the master")
			defer Conn.Close()

			//msg.Timestamp = msg.Timestamp+1
			/* Loads the buffer with the message in json-format */
			/*buf,err := json.Marshal(elev_system)
			CheckError(err)

			/* Sends the message */
			/*Conn.Write(buf)
			/* Uses sleep as a Timer for how long we will wait for the Slave to receive and process new orders*/
			/*time.Sleep(time.Second * 5)

		}else{
			 //Send more messages if problems
				MasterAddr, err := net.ResolveUDPAddr("udp", msg.MasterIP+"."strconv.Itoa(transmitPort))
				CheckError(err)

				/* Create a connection to the server */
				/*Conn, err := net.DialUDP("udp", nil, MasterAddr)
				CheckError(err)

				fmt.Println("This is a slave")
				defer Conn.Close()
				msg.Timestamp = broadcastElevator_System.Timestamp

				/* Loads the buffer with the message in json-format */
				/*buf,err := json.Marshal(msg)
				CheckError(err)

				Conn.Write(buf)	
		
			}
			
		}
		
	}
	
}*/

func UDPSendToSlave(transmitPort int, broadcastElevator_System chan Elevator_System ){
	e_system := <-broadcastElevator_System
	BroadcastAddr, err := net.ResolveUDPAddr("udp", "255.255.255.255:"+strconv.Itoa(transmitPort))
	CheckError(err)

	/* Create a connection to the server */
	Conn, err := net.DialUDP("udp", nil, BroadcastAddr)
	CheckError(err)

	fmt.Println("This is the master")
	defer Conn.Close()

	//msg.Timestamp = msg.Timestamp+1
	/* Loads the buffer with the message in json-format */
	buf,err := json.Marshal(e_system)
	CheckError(err)

	/* Sends the message */
	Conn.Write(buf)
	 
}

func UDPSendToMaster(transmitPort int, broadcastMessage chan Message){
	
	msg := <- broadcastMessage
	MasterAddr, err := net.ResolveUDPAddr("udp", msg.masterIP+"."+strconv.Itoa(transmitPort))
	CheckError(err)

	/* Create a connection to the server */
	Conn, err := net.DialUDP("udp", nil, MasterAddr)
	CheckError(err)

	fmt.Println("This is a slave")
	defer Conn.Close()
	//msg.timestamp = broadcastMessage.timestamp

	/* Loads the buffer with the message in json-format */
	buf,err := json.Marshal(msg)
	CheckError(err)

	Conn.Write(buf)	
}