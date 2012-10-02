package main

import ("os"
	"net"
	"math/rand"
	"time"
	"flag"
	"mud"
	"fmt")

type NamePrompt struct {
	mud.ConnectionState
	universe *mud.Universe
	StartRoom *mud.Room
	PlayerRemoveChan chan *mud.Player
}

func (n *NamePrompt) Name() string { return "name prompt" }
func (n *NamePrompt) Init(c *mud.UserConnection) {
	c.Write(Preamble)
	c.Write("Welcome. Please enter your name:\n\r")
}
func (n *NamePrompt) Respond(c *mud.UserConnection) bool {
	playerName := <- c.FromUser
	c.Data["playerName"] = playerName

	newP := n.universe.PlayerFromUserConn(c)
	mud.PlacePlayerInRoom(n.StartRoom, newP)
	mud.Look(newP, []string{})
	
	go newP.ReadLoop(n.PlayerRemoveChan)
	go newP.ExecCommandLoop()
	go mud.StimuliLoop(newP)
	
	return false
}

func main() {
	flagPort := flag.Int("port", 3000,
		"port to listen for mud clients")
	flagUseSeed := flag.Bool("seed", false, 
		"flush DB and seed universe with prototype's seed.go")
	flagUseLoad := flag.Bool("load", true, 
		"load objects from DB")
	flagSpeedupFactor := flag.Float64("speedup", 1.0,
		"factor to speed up heartbeat loop (2.0 means heartbeats come twice as often)")
	flag.Usage = func() {
		flag.PrintDefaults()
	}
	flag.Parse()
	mud.Log("program args: ", os.Args)

	rand.Seed(time.Now().Unix())
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d",*flagPort))
	universe := mud.NewUniverse()
	universe.Maker = BuildFFInRoom
	playerRemoveChan := make(chan *mud.Player)

	var theRoom *mud.Room
	if *flagUseSeed {
		mud.Log("Seeding Universe")
		universe.ClearDB()
		theRoom = InitUniverse(universe)
	} else if *flagUseLoad {
		mud.Log("Loading Universe")
		theRoom = LoadUniverse(universe)
		mud.Log("theRoom",theRoom)
	}

	mud.Log("len(rooms) =",len(universe.Rooms))

	go universe.HandlePersist()
	go universe.HeartbeatLoop(*flagSpeedupFactor)

	if err == nil {
		go mud.PlayerListManager(playerRemoveChan, universe.Players)
		defer listener.Close()

		mud.Log("Listening on port", *flagPort)
		for {
			conn, aerr := listener.Accept()
			if aerr == nil {
				namePrompt := new(NamePrompt)
				namePrompt.universe = universe
				namePrompt.StartRoom = theRoom
				namePrompt.PlayerRemoveChan = playerRemoveChan
				mud.NewUserConnection(conn, namePrompt)
			} else {
				mud.Log("Error in accept")
				mud.Log(aerr)
			}
		}
	} else {
		mud.Log("Error in listen", err)
	}
}