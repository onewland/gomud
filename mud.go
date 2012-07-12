package main

import ("fmt"
	"net"
	"math/rand"
	"strconv"
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

type HeartbeatClock struct { 
	mud.PhysicalObject
	mud.TimeListener
	counter int
	tPing chan int
}

func (c HeartbeatClock) Visible() bool { return true }
func (c HeartbeatClock) Carryable() bool { return false }
func (c HeartbeatClock) TextHandles() []string { 
	return []string{"clock", "heartbeat clock"} 
}
func (c HeartbeatClock) Description() string { 
	return "A large clock reading " + strconv.Itoa(c.counter)
}
func (c *HeartbeatClock) Ping() chan int { return c.tPing }

func (c *HeartbeatClock) UpdateTimeLoop() {
	for { c.counter = <- c.tPing }
}

func MakeClock() *HeartbeatClock {
	clock := new(HeartbeatClock)
	clock.tPing = make(chan int)
	return clock
}

func MakeStupidRoom() *mud.Room {
	theBall := Ball{}
	theClock := MakeClock()
	mud.TimeListenerList = []mud.TimeListener{theClock}
	ballSlice := []mud.PhysicalObject{theBall, theClock}
	empty := []mud.PhysicalObject{}

	room := mud.NewBasicRoom(1, "You are in a bedroom.", ballSlice)
	room2 := mud.NewBasicRoom(2, "You are in a bathroom.", empty)

	mud.ConnectEastWest(room, room2)

	go room.FanOutBroadcasts()
	go room2.FanOutBroadcasts()
	go room.ActionQueue()
	go room2.ActionQueue()
	go theClock.UpdateTimeLoop()

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

	go HeartbeatLoop(mud.TimeListenerList)

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

func HeartbeatLoop(listeners []mud.TimeListener) {
	for n:=0 ; ; n++ {
		for _, l := range(listeners) {
			l.Ping() <- n
		}
		time.Sleep(1*time.Millisecond)
	}
}