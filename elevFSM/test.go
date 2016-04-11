package main

import (
	"fmt"
	"sync"
	"time"
)

type Receiving_Channel struct {
	mtx   *sync.RWMutex
	value int
}

func Write_channel_value(recCh *Receiving_Channel) {
	recCh.mtx.Lock()
	recCh.value = 1
	time.Sleep(time.Second * 2)
	recCh.mtx.Unlock()
	fmt.Println(recCh.value)
}

func main() {

	locCh := new(Receiving_Channel)
	go Write_channel_value(&locCh)
	time.Sleep(time.Millisecond * 100)
	locCh.value = 42
	fmt.Println("Tried to write 42 to locCh, value: ", locCh.value)
	time.Sleep(time.Second * 3)
	fmt.Println("Tried to write 42 to locCh, value: ", locCh.value)
}
