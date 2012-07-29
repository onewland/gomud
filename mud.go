package main

import ("fmt"
	"net"
	"math/rand"
	"time"
	"mud")

func main() {
	rand.Seed(time.Now().Unix())
	listener, err := net.Listen("tcp", ":3000")
	universe := mud.NewBasicUniverse()
	universe.Maker = BuildFFInRoom
	playerRemoveChan := make(chan *mud.Player)
	idGen := UniqueIDGen()
	theRoom := MakeStupidRoom(universe)
	fmt.Println("len(rooms) =",len(universe.Rooms))

	go universe.HandlePersist()
	go HeartbeatLoop(universe.TimeListeners)

	if err == nil {
		go mud.PlayerListManager(playerRemoveChan, universe.Players)
		defer listener.Close()

		fmt.Println("Listening on port 3000")
		for {
			conn, aerr := listener.Accept()
			if aerr == nil {
				newP := universe.AcceptConnAsPlayer(conn, idGen)

				mud.PlacePlayerInRoom(theRoom, newP)

				go newP.ReadLoop(playerRemoveChan)
				go newP.ExecCommandLoop()
				go mud.StimuliLoop(newP)
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