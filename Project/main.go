package main

import (
	. "./elevController"
	"time"
)

func main() {
	/* INITIALIZATION */
	FSM_setup_elevator()

	/* SETS INITIAL STATE VARIABLES */
	Orders_init()
	e := FSM_create_elevator()
	e_system := Initialize_elev_system(&e)

	/* CHANNELS FOR UPDATING THE ELEVATOR VARIABLES */
	Button_Press_Chan := make(chan Button, 10)
	Location_Chan := make(chan int, 1)
	Motor_Direction_Chan := make(chan int, 1)
	Destination_Chan := make(chan int, 1)
	State_Chan := make(chan int, 1)

	/* EVENT CHANNELS */
	Objective_Chan := make(chan Button, 1)
	Floor_Arrival_Chan := make(chan int, 1)
	Door_Open_Req_Chan := make(chan int, 1)
	Door_Close_Req_Chan := make(chan int, 1)

	/* MESSAGE CHANNELS */
	Rchv_Message_Chan := make(chan Message, 10)
	Broadcast_Message_Chan := make(chan Message, 1)
	//Master_Ready_To_Send_Chan := make(chan bool)
	//Broadcast_Elev_System_Chan := make(chan Elevator_System, 1)
	//Rchv_Elev_System_Chan := make(chan Elevator_System, 1)
	//To_Master_Chan := make(chan Message)

	/* MASTER CHANNELS */
	Ping_Slaves_Chan := make(chan string, 1)
	Time_Window_Timeout_Chan := make(chan int, 1)
	//Master_Req_Update_Chan := make(chan bool, 1)
	From_Master_NewUpdate_Chan := make(chan Message, 10)
	From_Master_ReqSys_Chan := make(chan int, 1)


	/* STARTS ESSENTIAL PROCESSES */
	go Order_handler(Button_Press_Chan)
	go Get_internal_orders(&e, &e_system)
	go FSM_safekill()
	go FSM_sensor_pooler(Button_Press_Chan)
	go FSM_floor_tracker(&e, Location_Chan, Floor_Arrival_Chan)
	go FSM_objective_dealer(&e, State_Chan, Destination_Chan, Objective_Chan)
	go FSM_elevator_updater(&e, Motor_Direction_Chan, Location_Chan, Destination_Chan, State_Chan)

	/* STARTS THE NETWORK BETWEEN THE ELEVATORS AND THE MESSAGE-PASSING */
	go UDPListenForPing(PINGPORT, e_system, From_Master_ReqSys_Chan)// used PORT earlier
	go UDPListenForUpdateSlave(SUPDATEPORT, &e_system, From_Master_NewUpdate_Chan)// used PORT earlier
	if Is_elev_master(e_system) {
		go UDPListenForUpdateMaster(MUPDATEPORT, Rchv_Message_Chan)// used PORT earlier
		Ping_Slaves_Chan <- "1" //Initiates the master events
	}

	// Channels to see if slaves are alive
	



	time.Sleep(time.Millisecond * 200)

	Print_all_orders()

	for {
		select {
		/* FSM EVENTS: */
		case newObjective := <-Objective_Chan:
			FSM_Start_Driving(newObjective, &e, State_Chan, Motor_Direction_Chan, Location_Chan)

		case newFloorArrival := <-Floor_Arrival_Chan:
			FSM_should_stop_or_not(newFloorArrival, &e, State_Chan, Motor_Direction_Chan, Door_Open_Req_Chan)

		case doorOpenReq := <-Door_Open_Req_Chan:
			go FSM_door_opener(doorOpenReq, Door_Close_Req_Chan, State_Chan)

		case doorCloseReq := <-Door_Close_Req_Chan:
			FSM_door_closer(doorCloseReq, &e, State_Chan)

		/* NETWORK EVENTS: */
		/* MASTER ONLY EVENTS */
		case sendReq := <-Ping_Slaves_Chan:
			UDPSendReqToSlaves(PINGPORT, sendReq)            //Ping slaves for them to send their system info. Used PINGPORT earlier
			go Int_Timer_Chan(Time_Window_Timeout_Chan, 50) //Opens a time window

		case infoRec := <-Rchv_Message_Chan:
			Message_Compiler_Master(infoRec, &e_system) //Gathering system info meanwhile

		case <-Time_Window_Timeout_Chan: //Time window closes, starts processing info
			//fmt.Printf("\nNETWORK: SYSTEM INFO GATHERED, PROCESSING - %d\n", timeWindowTimeout)
			UDPSendSysInfoToSlaves(SUPDATEPORT, e_system) //Start processing information gathered, then send it. Used SUPDATEPORT earlier
			go String_Timer_Chan(Ping_Slaves_Chan, 400)   //Sends it out and waits before it opens a new time window

		/* SLAVE EVENTS */
		case <-From_Master_ReqSys_Chan:
			//fmt.Printf("\nNETWORK: SYSTEM INFO REQUEST FROM MASTER - %d\n", sendMyInfo)
			MessageSetter(Broadcast_Message_Chan, e_system, &e)
			UDPSendToMaster(MUPDATEPORT, Broadcast_Message_Chan) //Used MUPDATEPORT earlier

		case newSysInfo := <-From_Master_NewUpdate_Chan:
			//fmt.Println("\nNETWORK: RECIEVED NEW SYSTEM INFO FROM MASTER")
			Sync_with_system(newSysInfo, &e, &e_system)
		/*
		case aliveReq := <- Alive_Ping_Chan:
			UDPSendAliveMessage(blabla)
		*/
		}
	}
}
