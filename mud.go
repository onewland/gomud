package main

import ("fmt"
	"net"
	"math/rand"
	"strconv"
	"strings"
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
func (b Ball) Description() string { return "A &red;red&; ball" }
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

type FlipFlop struct {
	mud.NPC
	mud.Persister
	room *mud.Room
	stimuli chan mud.Stimulus
	id int
	lastText string
}

func (f FlipFlop) ID() int { return f.id }
func (f FlipFlop) Name() string { return f.lastText }
// Only respond to Talk stimulus to copy them
func (f FlipFlop) DoesPerceive(s mud.Stimulus) bool {
	_, ok := s.(mud.TalkerSayStimulus)
	return ok
}
func (f FlipFlop) TextHandles() []string { return []string{} }
func (f *FlipFlop) HandleStimulus(s mud.Stimulus) {
	scast, ok := s.(mud.TalkerSayStimulus)
	if !ok {
		panic("Puritan should only receive TalkerSayStimulus")
	} else {
		args := strings.SplitN(scast.Text()," ",3)
		fmt.Println("FF args:",args)
		if(args[0] == "bling") {
			switch(args[1]) {
			case "set":
				f.lastText = args[2]
			}
		}
	}
}
func (f FlipFlop) StimuliChannel() chan mud.Stimulus {
	return f.stimuli
}
func (f FlipFlop) Visible() bool { return true }
func (f FlipFlop) Description() string { return f.Name() }
func (f FlipFlop) Carryable() bool { return false }
func (f FlipFlop) Save() (string, bool) {
	finished := false
	return "x",finished
}

type Puritan struct {
	mud.Talker
	mud.NPC
	room *mud.Room
	stimuli chan mud.Stimulus
	id int
}

func (p Puritan) ID() int { return p.id }
func (p Puritan) Name() string { return "Mary Magdalene" }
// Only respond to Talk stimulus to scorn people for cursing
func (p Puritan) DoesPerceive(s mud.Stimulus) bool {
	_, ok := s.(mud.TalkerSayStimulus)
	return ok
}
func (p Puritan) TextHandles() []string {
	return []string { "mary", "mm" }
}

func ContainsAny(s string, subs ...string) bool {
	for _,sub := range(subs) {
		if(strings.Contains(s, sub)) {
			return true
		}
	}
	return false
}

func (p Puritan) HandleStimulus(s mud.Stimulus) {
	scast, ok := s.(mud.TalkerSayStimulus)
	stim := mud.TalkerSay(p, "Wash your mouth out, " + scast.Source().Name())
	if !ok {
		panic("Puritan should only receive TalkerSayStimulus")
	} else {
		text := scast.Text()
		if(ContainsAny(text,
			"shit","piss","fuck",
			"cunt","cocksucker",
			"motherfucker","tits")) {
			p.room.Broadcast(stim)
		}
	}
}
func (p Puritan) StimuliChannel() chan mud.Stimulus {
	return p.stimuli
}
func (p Puritan) Visible() bool { return true }
func (p Puritan) Description() string { return p.Name() }
func (p Puritan) Carryable() bool { return false }

func MakeFlipFlop() *FlipFlop {
	ff := new(FlipFlop)
	ff.id = 101
	ff.lastText = "Unchanged."
	ff.stimuli = make(chan mud.Stimulus, 5)
	return ff
}

func MakePuritan() *Puritan {
	puritan := new(Puritan)
	puritan.id = 100
	puritan.stimuli = make(chan mud.Stimulus, 5)
	return puritan
}

func MakeClock() *HeartbeatClock {
	clock := new(HeartbeatClock)
	clock.tPing = make(chan int)
	return clock
}

func MakeStupidRoom(universe *mud.Universe) *mud.Room {
	puritan := MakePuritan()
	ff := MakeFlipFlop()
	theBall := Ball{}
	theClock := MakeClock()
	universe.TimeListeners = []mud.TimeListener{theClock}
	ballSlice := []mud.PhysicalObject{theBall, theClock, puritan, ff}
	empty := []mud.PhysicalObject{}

	room := mud.NewBasicRoom(universe, 1, "You are in a bedroom.", ballSlice)
	room.AddPerceiver(puritan)
	room.AddPerceiver(ff)
	room2 := mud.NewBasicRoom(universe, 2, "You are in a bathroom.", empty)
	puritan.room = room
	go mud.StimuliLoop(puritan)
	go mud.StimuliLoop(ff)

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
	universe := mud.NewBasicUniverse()
	playerRemoveChan := make(chan *mud.Player)
	idGen := UniqueIDGen()
	theRoom := MakeStupidRoom(universe)
	fmt.Println("len(rooms) =",len(universe.Rooms))

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