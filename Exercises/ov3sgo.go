package main

// ip-adress: 129.241.187.23
import (
	"fmt"
	"net"
	"os"
	"runtime"
	"strconv"
	"time"
)

/* A Simple function to verify error */
func CheckError(err error) {
	if err != nil {
		fmt.Println("Error: ", err)
		os.Exit(0)
	}
}

func serverListen() {

	/* Lets prepare a address at any address at port 30000*/
	/* For testing: sett addresse lik ip#255:30000*/
	ServerAddr, err := net.ResolveUDPAddr("udp", ":30000")
	CheckError(err)

	/* Now listen at selected port */
	ServerConn, err := net.ListenUDP("udp", ServerAddr)
	CheckError(err)
	defer ServerConn.Close()

	buffer := make([]byte, 1024)

	for {
		n, addr, err := ServerConn.ReadFromUDP(buffer)
		fmt.Println("Received ", string(buffer[0:n]), " from ", addr)

		if err != nil {
			fmt.Println("Error: ", err)
		}
		time.Sleep(time.Second * 1)
	}
}

func serverSend() {

	/* Dial up UDP */
	LocalAddr, err := net.ResolveUDPAddr("udp", "129.241.187.255:20020")
	CheckError(err)
	// making a connection to server
	Conn, err := net.DialUDP("udp", nil, LocalAddr)
	CheckError(err)

	defer Conn.Close()
	i := 1337
	for {
		msg := strconv.Itoa(i)
		i++
		buf := []byte(msg)
		_, err := Conn.Write(buf)
		if err != nil {
			fmt.Println(msg, err)
		}
		time.Sleep(time.Second * 5)
	}
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	go serverListen()
	go serverSend()
	time.Sleep(time.Second * 30)
}
