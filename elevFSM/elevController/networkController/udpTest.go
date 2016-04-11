package main

// ip-adress: 129.241.187.23
import (
	"fmt"
	"net"
	"os"
	"runtime"
	"time"
	"encoding/json"
	"strconv"
)


const PORT int = 40000;//setting the portnumber. Selected a random port for testing purposes
type Message struct {
	Target_Floor int
	Current_Floor_location int 
	Id int
	Timestamp int
}


/* A Simple function to verify error */
func CheckError(err error) {
	if err != nil {
		fmt.Println("Error: ", err)
		os.Exit(0)
	}
}

func Message_reader(storage_channel chan Message){
	for{
		Message := <- storage_channel
		fmt.Println(Message)
	}
}


func UDPListen(isMaster bool, listenPort int, masterIP string) {

	/* For testing: sett addresse lik ip#255:30000*/
	//ServerAddr, err := net.ResolveUDPAddr("udp", ":40000")
	ServerAddr, err := net.ResolveUDPAddr("udp", ":"+strconv.Itoa(listenPort))
	CheckError(err)

	/* Now listen at selected port */
	ServerConn, err := net.ListenUDP("udp", ServerAddr)
	CheckError(err)
	defer ServerConn.Close()
	var received_message Message 

	storageChannel := make(chan Message)
	buffer := make([]byte, 1024)
	trimmed_buffer:= make([]byte, 1024)
	go Message_reader(storageChannel)
		if (isMaster){
			fmt.Println("I am master")
			for {
					n, addr, err := ServerConn.ReadFromUDP(buffer)
					trimmed_buffer=buffer[0:n]
					fmt.Println("Received ", string(buffer[0:n]), " from ", addr)
					CheckError(err)
					err= json.Unmarshal(trimmed_buffer,&received_message)
					CheckError(err)
					//fmt.Println(received_message)
					storageChannel <- received_message
				}	
		}else{
			for {
				n, addr, err := ServerConn.ReadFromUDP(buffer)
				if (addr.String()==masterIP){
					trimmed_buffer=buffer[0:n]
					fmt.Println("Received ", string(buffer[0:n]), " from ", addr)
					CheckError(err)
					fmt.Println("Unmarshal")
					err= json.Unmarshal(trimmed_buffer,&received_message)
					CheckError(err)
					storageChannel <- received_message
					time.Sleep(time.Second * 1)
				}
			
			}
		}
	
	}



 
//need to include message-sending
func UDPSend(isMaster bool,transmitPort int,masterIP string, broadcastChann chan Message) {

	/* Dial up UDP */

	if (isMaster){
		BroadcastAddr, err := net.ResolveUDPAddr("udp", "255.255.255.255:"+strconv.Itoa(transmitPort))
		CheckError(err)
		/* Make a connection to the server */
		Conn, err := net.DialUDP("udp", nil, BroadcastAddr)
		CheckError(err)
		fmt.Println("This is the master")
		defer Conn.Close()
		/*msg1:=&Message{
		 Target_Floor: 3,
		 Current_Floor_location: 2,
		 Id: 1,
		 Timestamp: 0}
		*/
		for {
		//msg1.Timestamp = msg1.Timestamp+1
		//buf,err := json.Marshal(msg1)
		buf,err := json.Marshal(<-broadcastChann)
		CheckError(err)
		Conn.Write(buf)
		time.Sleep(time.Second * 5)
		}

	} else {
		MasterAddr, err := net.ResolveUDPAddr("udp", masterIP+"+" + strconv.Itoa(transmitPort))
		CheckError(err)
		/* Make a connection to the server */
		Conn, err := net.DialUDP("udp", nil, MasterAddr)
		CheckError(err)
		fmt.Println("This is a slave")

		defer Conn.Close()
		msg1:=&Message{
		 Target_Floor: 3,
		 Current_Floor_location: 2,
		 Id: 1,
		 Timestamp: 0}
	
		for {
		msg1.Timestamp = msg1.Timestamp+1
		buf,err := json.Marshal(msg1)
		CheckError(err)
		Conn.Write(buf)

		time.Sleep(time.Second * 5)
	
		}

	}

	
	
}

	
func testChannels(broadcastChann chan Message){
	for{
		msg1:=Message{
			 Target_Floor: 3,
			 Current_Floor_location: 2,
			 Id: 1,
			 Timestamp: 0}
			 broadcastChann <- msg1
	}
}
	

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	broadcastChann := make(chan Message)

	go testChannels(broadcastChann)
	go UDPListen(true,PORT,"129.241.187.156")
	go UDPSend(true,PORT,"129.241.187.156",broadcastChann)
	time.Sleep(time.Second * 60)
}
