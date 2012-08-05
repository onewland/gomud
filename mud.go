package main

import ("os"
	"net"
	"math/rand"
	"time"
	"flag"
	"mud")


func main() {
	flagUseSeed := flag.Bool("seed", false, 
		"flush DB and seed universe with prototype's seed.go")
	flagUseLoad := flag.Bool("load", true, 
		"load objects from DB")
	flag.Usage = func() {
		flag.PrintDefaults()
	}
	flag.Parse()
	mud.Log("program args: ", os.Args)

	rand.Seed(time.Now().Unix())
	listener, err := net.Listen("tcp", ":3000")
	universe := mud.NewBasicUniverse()
	universe.Maker = BuildFFInRoom
	playerRemoveChan := make(chan *mud.Player)
	idGen := UniqueIDGen()

	var theRoom *mud.Room
	if(*flagUseSeed) {
		mud.Log("Seeding Universe")
		universe.ClearDB()
		theRoom = MakeStupidRooms(universe)
	} else if(*flagUseLoad) {
		mud.Log("Loading Universe")
		theRoom = LoadStupidRooms(universe)
		mud.Log("theRoom",theRoom)
	}

	mud.Log("len(rooms) =",len(universe.Rooms))

	go universe.HandlePersist()
	go HeartbeatLoop(universe.TimeListeners)

	if err == nil {
		go mud.PlayerListManager(playerRemoveChan, universe.Players)
		defer listener.Close()

		mud.Log("Listening on port 3000")
		for {
			conn, aerr := listener.Accept()
			if aerr == nil {
				newP := universe.AcceptConnAsPlayer(conn, idGen)

				mud.PlacePlayerInRoom(theRoom, newP)

				go newP.ReadLoop(playerRemoveChan)
				go newP.ExecCommandLoop()
				go mud.StimuliLoop(newP)
			} else {
				mud.Log("Error in accept")
				mud.Log(aerr)
			}
		}
	} else {
		mud.Log("Error in listen")
		mud.Log(err)
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