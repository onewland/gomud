package main

import ("fmt"
	"net"
	"math/rand"
	"time"
	"mud")

/*
 * Beginnings of a MU* type server
 * 
 * Initially supporting: 
 * 1. Players - who can see each other, talk to each other
 * 2. Objects - visible or not, carry-able or not
 * 3. Simple commands (look, take)
 * 4. Rooms
 * 5. Heartbeat
 *
 * Initially not supporting:
 * - Persistence of state
 * - Combat
 * - NPCs
 */

type Ball struct { mud.PhysicalObject }

func (b Ball) Visible() bool { return true }
func (b Ball) Description() string { return "A red ball" }
func (b Ball) Carryable() bool { return true }
func (b Ball) TextHandles() []string { return []string{"ball","red ball"} }

func MakeStupidRoom() *mud.Room {
	theBall := Ball{}
	ballSlice := []mud.PhysicalObject{theBall}
	empty := []mud.PhysicalObject{}

	room := mud.NewBasicRoom(1, "You are in a bedroom.", ballSlice)
	room2 := mud.NewBasicRoom(2, "You are in a bathroom.", empty)

	mud.ConnectEastWest(room, room2)

	go room.FanOutBroadcasts()
	go room2.FanOutBroadcasts()
	go room.ActionQueue()
	go room2.ActionQueue()

	return room
}

func main() {
	rand.Seed(time.Now().Unix())
	listener, err := net.Listen("tcp", ":3000")
	playerRemoveChan := make(chan *mud.Player)
	mud.PlayerList = make(map[int]*mud.Player)
	mud.RoomList = make(map[mud.RoomID]*mud.Room)
	idGen := UniqueIDGen()
	theRoom := MakeStupidRoom()

	go HeartbeatLoop()

	if err == nil {
		go mud.PlayerListManager(playerRemoveChan, mud.PlayerList)
		defer listener.Close()

		fmt.Println("Listening on port 3000")
		for {
			conn, aerr := listener.Accept()
			if aerr == nil {
				newP := mud.AcceptConnAsPlayer(conn, idGen)

				mud.PlacePlayerInRoom(theRoom, newP)

				go newP.ReadLoop(playerRemoveChan)
				go newP.ExecCommandLoop()
				go newP.StimuliLoop()
			} else {
				fmt.Println("Error in accept")
			}
		}
	} else {
		fmt.Println("Error in listen")
	}
}

func UniqueIDGen() func() int {
	x, xchan := 0, make(chan int)
	go func() { for ; ; x += 1 { xchan <- x } }()

	return func() int { return <- xchan }
}

func HeartbeatLoop() {
	//recurScheduled := make(map[int]list.List)
	for n:=0 ; ; n++ {
		//fmt.Println("Heartbeat", time.Now())
		time.Sleep(5*time.Second)
	}
}