package main

import (
	"fmt"
	"net"
	"os/exec"
	"strconv"
	"time"
)

func CheckError(err error) {
	if err != nil {
		fmt.Println("Error: ", err)
	}
}

func master(ipAdr string, port string, counter int) {
	fmt.Println("This is the master now")

	serverAddr, err := net.ResolveUDPAddr("udp", ipAdr+":"+port)
	conn, err := net.DialUDP("udp", nil, serverAddr)

	CheckError(err)

	countSendS := ""
	countSend := make([]byte, 1024)

	time.Sleep(1 * time.Second)

	for {
		counter++
		countSendS = strconv.Itoa(counter)
		countSend = []byte(countSendS)
		_, err = conn.Write(countSend)

		CheckError(err)

		fmt.Print(counter)
		time.Sleep(1 * time.Second)

	}
}

func backup(ipAdr string, port string, counter int) {

	fmt.Println("This is the backup")
	serverAddr, err := net.ResolveUDPAddr("udp", ipAdr+":"+port)
	psock, err := net.ListenUDP("udp4", serverAddr)

	defer psock.Close()

	CheckError(err)

	buf := make([]byte, 1024)

	for {
		psock.SetDeadline(time.Now().Add(3 * time.Second))
		_, _, err := psock.ReadFromUDP(buf)

		if err != nil {
			Backup := exec.Command("gnome-terminal", "-x", "sh", "-c", "go run ex6.go")
			Backup.Run()
			return

		}
	}

}

func main() {
	counter := 0
	backup("localhost", "30000", counter)
	master("localhost", "30000", counter)
}
