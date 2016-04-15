//need to include message.go
package elevController

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	PINGPORT    int = 34014
	SUPDATEPORT int = 33014
	MUPDATEPORT int = 35014
	//PORT int = 30231

)

type Elevator_System struct {
	selfID    string
	selfIP    string
	//elevators map[int]*Elevator //elevator declared in FSM
	elevators map[string]*Elevator 	
	masterID    int
	masterIP  string
	timestamp int
}

type Message struct {
	destinationFloor int
	currentFloor     int
	ID               string
	timestamp        int
	InternalOrders   [10]Button
	master           bool
	masterIP         string
}

func Initialize_elev_system(elevator *Elevator) Elevator_System {
	var e_system Elevator_System
	e_system.elevators = make(map[string]*Elevator)
	addr, _ := net.InterfaceAddrs()
	tempVar := addr[1]
	e_system.timestamp = 1
	ip := tempVar.String()
	e_system.selfIP = ip[0:15]
	e_system.selfID = strconv.Itoa(int(addr[1].String()[12]-'0')*100 + int(addr[1].String()[13]-'0')*10 + int(addr[1].String()[14]-'0')) //this will work for IP-addresses of format ###.###.###.###, but not with only for ###.###.###.##/
	e_system.elevators[e_system.selfID] = new(Elevator)
	e_system.elevators[e_system.selfID].InternalOrders = elevator.InternalOrders
	fmt.Println("\nelevator.InternalOrders: ", elevator.InternalOrders)
	fmt.Println("\ne_system.elevators[e_system.selfID].InternalOrders: ", e_system.elevators[e_system.selfID].InternalOrders)

	//before initializing we need to actually listen if we receive something
	masterExists:=UDPListenForMasterInit(SUPDATEPORT,&e_system)
	fmt.Println("Does it exists a master: " + strconv.FormatBool(masterExists))
	if (masterExists==false){
		Set_master(&e_system)
	}
	
	//Checking Elev_system variables
	fmt.Println("Master is: " + strconv.Itoa(e_system.masterID))
	fmt.Printf("Self ID is: %s \n", e_system.selfID)
	fmt.Printf("Timestamp is: %d \n", e_system.timestamp)
	fmt.Printf("MasterIP is:" + e_system.masterIP)
	return e_system
}

func Is_elev_master(e_system Elevator_System) bool {
	isMaster := false
	if e_system.selfID == strconv.Itoa(e_system.masterID) {
		isMaster = true
	}
	return isMaster
}

func MessageSetter(Broadcast_Message_Chan chan Message, e_system Elevator_System, e *Elevator) {
	var msg Message
	msg.destinationFloor = e.DestinationFloor
	msg.currentFloor = e.CurrentFloor
	msg.ID = e_system.selfID
	msg.masterIP = e_system.masterIP
	msg.timestamp = e_system.timestamp
	msg.InternalOrders = e_system.elevators[e_system.selfID].InternalOrders
	//fmt.Println(msg.ID)
	if e_system.selfID == strconv.Itoa(e_system.masterID){
		msg.master = true
	} else {
		msg.master = false
	}
	Broadcast_Message_Chan <- msg
}

func Message_Compiler_Master(msgFromSlave Message, e_system *Elevator_System) {
	//fmt.Println("\nStarted compiling message")
	//fmt.Println("\nMessage: ", msgFromSlave)

	var elevExistedInMap bool=false
	for i,_ := range e_system.elevators{
		if i == msgFromSlave.ID {
			e_system.elevators[i].InternalOrders = msgFromSlave.InternalOrders
			elevExistedInMap=true
			break
		}
	}
	if elevExistedInMap==false{
		//fmt.Println("\nTrying to create a new elevator...\n")
		e_system.elevators[msgFromSlave.ID] = new(Elevator)
		//fmt.Println("\nCreated a new elevator...\n")
		e_system.elevators[msgFromSlave.ID].InternalOrders = msgFromSlave.InternalOrders
		//fmt.Println("\nTried to update its orders\n")
	}
}

func Remove_elev(ID string, e *Elevator_System) {
	delete(e.elevators, ID)
	fmt.Println("Elevator ", ID, " removed from network")
	Set_master(e)
}

func Add_elev(ID string, e_system *Elevator_System){
	e_system.elevators[ID] = new(Elevator)
}

func Set_master(e_system *Elevator_System) {
	// Checking which elevator has the highest IP to determine who is the master
	max := 0
	for i, _ := range e_system.elevators {
		j,_:=strconv.Atoi(i)
		if max < j {
			max = j

		}
	}
	e_system.masterID = max
	e_system.timestamp = 1
	var tempIP string = e_system.selfIP[0:12]
	e_system.masterIP = tempIP + strconv.Itoa(e_system.masterID)
	fmt.Println("New master is", e_system.masterID)
}

func Int_Timer_Chan(Timer_Chan chan int, n int) {
	timer := time.NewTimer(time.Millisecond * time.Duration(n))
	<-timer.C
	//fmt.Println("Timer timeout")
	Timer_Chan <- 1
}

func String_Timer_Chan(Timer_Chan chan string, n int) {
	timer := time.NewTimer(time.Millisecond * time.Duration(n))
	<-timer.C
	//fmt.Println("Timer timeout")
	Timer_Chan <- "1"
}

/* A Simple function to verify error */
func CheckError(err error) {
	if err != nil {
		fmt.Println("Error: ", err)
		os.Exit(0)
	}
}

func UDPListenForMasterInit(listenPort int,e_system *Elevator_System) bool{
	ServerAddr, err := net.ResolveUDPAddr("udp", ":"+strconv.Itoa(listenPort))
	CheckError(err)
	var masterExists bool 

	/* Now listen at selected port */
	ServerConn, err := net.ListenUDP("udp", ServerAddr)
	CheckError(err)
	defer ServerConn.Close()
	buf := make([]byte, 1024)
	ServerConn.SetDeadline(time.Now().Add(time.Millisecond*1000))
	_,addr, err1 := ServerConn.ReadFromUDP(buf)
		if err1 !=nil{
				masterExists=false
		}else{
			e_system.masterIP = strings.Split(addr.String(), ":")[0]
			masterID,_ := strconv.Atoi(e_system.masterIP[12:])//little bit insecure about the ParseInt function
			//fmt.Printf("\nDummyint contains: %d,",masterID)
			e_system.masterID = masterID
			masterExists = true
		}						//removed for-loop containing breaks;(if not responding)
	return masterExists

}

func UDPListenForPing(listenPort int, e_system Elevator_System, From_Master_ReqSys_Chan chan int) {

	//ServerAddr, err := net.ResolveUDPAddr("udp", ":40000")
	ServerAddr, err := net.ResolveUDPAddr("udp", ":"+strconv.Itoa(listenPort))
	CheckError(err)

	/* Now listen at selected port */
	ServerConn, err := net.ListenUDP("udp", ServerAddr)
	CheckError(err)
	defer ServerConn.Close()

	buffer := make([]byte, 1024)
	trimmed_buffer := make([]byte, 1)
	for { // what if the elevators crashes and the slave becomes the master
		n, addr, err := ServerConn.ReadFromUDP(buffer)
		//fmt.Printf("\naddr: %d", addr.String())
		//fmt.Printf("\ne_system.masterIP: %d", e_system.masterIP)
		if strings.Split(addr.String(), ":")[0] == e_system.masterIP {
			//fmt.Println("hei")
			trimmed_buffer = buffer[0:n]
			i := string(trimmed_buffer)
			//fmt.Println("i:", i, "  i == \"1\":", i == "1")
			if i == "1" {
				//fmt.Println("\nPing received ", i, " from ", addr)
				CheckError(err)

				//err = json.Unmarshal(trimmed_buffer, &received_message)
				From_Master_ReqSys_Chan <- 1
				time.Sleep(time.Millisecond * 2)
			}
		}
	}
}

func UDPListenForUpdateMaster(listenPort int, infoRec chan Message) {

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
	trimmed_buffer := make([]byte, 1)
	for {
		n, _, err := ServerConn.ReadFromUDP(buffer)
		trimmed_buffer = buffer[0:n]
		//fmt.Println("Received ", string(buffer[0:n]), " from ", addr)
		CheckError(err)

		err = json.Unmarshal(trimmed_buffer, &received_message)
		CheckError(err)

		infoRec <- received_message
		time.Sleep(time.Millisecond * 50)
	}

}

func UDPListenForUpdateSlave(listenPort int, e_system *Elevator_System, From_Master_NewUpdate_Chan chan Message) {

	//For testing: sett addresse lik ip#255:30000
	//ServerAddr, err := net.ResolveUDPAddr("udp", ":40000")
	ServerAddr, err := net.ResolveUDPAddr("udp", ":"+strconv.Itoa(listenPort))
	CheckError(err)

	// Now listen at selected port
	ServerConn, err := net.ListenUDP("udp", ServerAddr)
	CheckError(err)
	defer ServerConn.Close()

	var msg Message

	buffer := make([]byte, 1024)
	trimmed_buffer := make([]byte, 1)

	for { // what if the elevators crashes and the slave becomes the master
		n, addr, err := ServerConn.ReadFromUDP(buffer)
		if strings.Split(addr.String(), ":")[0] == e_system.masterIP{
			trimmed_buffer = buffer[0:n]
			//fmt.Println("Received ", string(buffer[0:n]), " from ", addr)
			CheckError(err)
			err = json.Unmarshal(trimmed_buffer, &msg)
			CheckError(err)
			From_Master_NewUpdate_Chan <- msg
		}
	}
}


func UDPSendReqToSlaves(transmitPort int, ping string) {
	BroadcastAddr, err := net.ResolveUDPAddr("udp", "255.255.255.255:"+strconv.Itoa(transmitPort))
	CheckError(err)

	/* Create a connection to the server */
	Conn, err := net.DialUDP("udp", nil, BroadcastAddr)
	CheckError(err)

	//fmt.Println("This is the master")
	defer Conn.Close()

	Conn.Write([]byte(ping))
}

func UDPSendSysInfoToSlaves(transmitPort int, reworkedSystem Elevator_System) {
	BroadcastAddr, err := net.ResolveUDPAddr("udp", "255.255.255.255:"+strconv.Itoa(transmitPort))
	CheckError(err)
	//fmt.Println("\n\n\n\n\nreworkedSystem.elevators: ", reworkedSystem.elevators[reworkedSystem.selfID].InternalOrders)

	/* Create a connection to the server */
	Conn, err := net.DialUDP("udp", nil, BroadcastAddr)
	CheckError(err)
	defer Conn.Close()
	var msg Message

	for i, _ := range reworkedSystem.elevators {
		msg.destinationFloor = reworkedSystem.elevators[i].DestinationFloor
		msg.currentFloor = reworkedSystem.elevators[i].CurrentFloor
		msg.ID = reworkedSystem.selfID
		msg.masterIP = reworkedSystem.masterIP
		msg.timestamp = reworkedSystem.timestamp
		msg.InternalOrders = reworkedSystem.elevators[i].InternalOrders
		buf, err := json.Marshal(msg)
		Conn.Write(buf)
		CheckError(err)

		//fmt.Println("\nMSG to buffer : ", string(buf))
	}
}



func UDPSendToMaster(transmitPort int, broadcastMessage_Chan chan Message) {
	msg := <-broadcastMessage_Chan
	//fmt.Println("\n MSG: ",msg,"\n")
	MasterAddr, err := net.ResolveUDPAddr("udp", msg.masterIP+":"+strconv.Itoa(transmitPort))
	CheckError(err)

	/* Create a connection to the server */
	Conn, err := net.DialUDP("udp", nil, MasterAddr)
	CheckError(err)

	//fmt.Println("This is a slave")
	defer Conn.Close()
	//msg.timestamp = broadcastMessage.timestamp

	/* Loads the buffer with the message in json-format */
	buf, err := json.Marshal(msg)
	CheckError(err)

	Conn.Write(buf)
}
