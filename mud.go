package main

import ("fmt"
	"net"
	"math/rand"
	"time"
	"io"
	"strings"
	"regexp"
)

/*
 * Beginnings of a MU* type server
 * 
 * Initially supporting: 
 * 1. Players - who can see each other, talk to each other
 * 2. Objects - visible or not, carry-able or not
 * 3. Simple commands (look, take)
 * 4. Rooms
 * 5. Heartbeat (?)
 *
 * Initially not supporting:
 * - Persistence of state
 * - Combat
 * - NPCs
 */

type PerceiveMap map[string]PhysicalObject

type Perceiver interface {
	DoesPerceive(s Stimulus) bool
	PerceiveList() PerceiveMap
}

type Stimulus interface {
	StimType() string
	Description(p Perceiver) string
}

type PlayerEnterStimulus struct {
	Stimulus
	player *Player
}

type PlayerLeaveStimulus struct {
	Stimulus
	player *Player
}

type PlayerSayStimulus struct {
	Stimulus
	player *Player
	text string
}

type PlayerPickupStimulus struct {
	Stimulus
	player *Player
	obj PhysicalObject
}

type RoomID int

type Room struct {
	id RoomID
	text string
	players []Player
	physObjects []PhysicalObject
	stimuliBroadcast chan Stimulus
}

type PhysicalObject interface {
	Visible() bool
	Description() string
	Carryable() bool
	TextHandles() []string
}

type Ball struct { PhysicalObject }

func (b Ball) Visible() bool { return true }
func (b Ball) Description() string { return "A red ball" }
func (b Ball) Carryable() bool { return true }
func (b Ball) TextHandles() []string { return []string{"ball","red ball"} }

type Player struct {
	Perceiver
	PhysicalObject
	id int
	room RoomID
	name string
	sock net.Conn
	inventory []PhysicalObject
	commandBuf chan string
	stimuli chan Stimulus
}

func (p Player) Visible() bool { return true }
func (p Player) Description() string { return "A person: " + p.name }
func (p Player) Carryable() bool { return false }
func (p Player) TextHandles() []string { return []string{ strings.ToLower(p.name) } }

var playerList map[int]*Player
var roomList map[RoomID]*Room

func MakeStupidRoom() *Room {
	room := Room{id: 1}
	room.text = "You are in a bedroom."
	room.stimuliBroadcast = make(chan Stimulus, 10)
	theBall := Ball{}
	room.physObjects = []PhysicalObject {theBall}
	go room.FanOutBroadcasts()
	return &room
}

func main() {
	rand.Seed(time.Now().Unix())
	listener, err := net.Listen("tcp", ":3000")
	playerRemoveChan := make(chan *Player)
	playerList = make(map[int]*Player)
	roomList = make(map[RoomID]*Room)
	idGen := UniqueIDGen()
	theRoom := MakeStupidRoom()

	roomList[theRoom.id] = theRoom

	if err == nil {
		go PlayerListManager(playerRemoveChan, playerList)
		defer listener.Close()

		fmt.Println("Listening on port 3000")
		for {
			conn, aerr := listener.Accept()
			if aerr == nil {
				newP := AcceptConnAsPlayer(conn, idGen)
				playerList[newP.id] = newP

				PlacePlayerInRoom(theRoom, newP)

				fmt.Println(newP.name, "joined, ID =",newP.id)
				fmt.Println(len(playerList), "player[s] online.")

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

func PlacePlayerInRoom(r *Room, p *Player) {
	oldRoomID := p.room
	if oldRoomID != -1 {
		oldRoom := roomList[oldRoomID]
		oldRoom.stimuliBroadcast <- 
			PlayerLeaveStimulus{player: p}
	}
	
	p.room = r.id
	r.stimuliBroadcast <- PlayerEnterStimulus{player: p}
	r.players = append(r.players, *p)
}

func UniqueIDGen() func() int {
	x, xchan := 0, make(chan int)

	go func() {
		for ; ; x += 1 { xchan <- x }
	}()

	return func() int { return <- xchan }
}

func PlayerListManager(toRemove chan *Player, pList map[int]*Player) {
	for {
		pRemove := <- toRemove
		delete(pList, pRemove.id)
		fmt.Println("Removed", pRemove.name, "from player list")
	}
}

func SplitCommandString(cmd string) []string {
	re, _ := regexp.Compile(`(\S+)|(['"].+['"])`)
	return re.FindAllString(cmd, 10)
}

func (p *Player) ExecCommandLoop() {
	for {
		nextCommand := <-p.commandBuf
		nextCommandSplit := SplitCommandString(nextCommand)
		if nextCommandSplit != nil && len(nextCommandSplit) > 0 {
			nextCommandRoot := nextCommandSplit[0]
			nextCommandArgs := nextCommandSplit[1:]
			fmt.Println("Next command from",p.name,":",nextCommandRoot)
			fmt.Println("args:",nextCommandArgs)
			if nextCommandRoot == "who" {
				p.Who(nextCommandArgs) 
			} else if nextCommandRoot == "look" {
				p.Look(nextCommandArgs) 
			} else if nextCommandRoot == "say" {
				p.Say(nextCommandArgs) 
			} else if nextCommandRoot == "take" {
				p.Take(nextCommandArgs) 
			}
		}
		p.WriteString("> ")
	}
}

func (r *Room) FanOutBroadcasts() {
	for {
		broadcast := <- r.stimuliBroadcast
		for _,p := range r.players { 
			p.stimuli <- broadcast 
		}
	}
}

func (p *Player) Look(args []string) {
	if len(args) > 1 {
		fmt.Println("Too many args")
		p.WriteString("Too many args")
	} else {
		p.WriteString(roomList[p.room].Describe(p) + "\n")
	}
}

func (p *Player) Who(args []string) {
	gotOne := false
	for id, pOther := range playerList {
		if id != p.id {
			str_who := fmt.Sprintf("[WHO] %s\n",pOther.name)
			p.WriteString(str_who)
			gotOne = true
		}
	}

	if !gotOne {
		p.WriteString("You are all alone in the world.\n")
	}
}

func (p *Player) Say(args []string) {
	room := roomList[p.room]
	sayStim := PlayerSayStimulus{player: p, text: strings.Join(args," ")}
	room.stimuliBroadcast <- sayStim
}

func (p *Player) Take(args []string) {
	room := roomList[p.room]
	if len(args) > 0 {
		target := strings.ToLower(args[0])

		if targetObj, ok := p.PerceiveList()[target]; ok {
			if targetObj.Carryable() {
				room.stimuliBroadcast <- PlayerPickupStimulus{player: p, obj: targetObj}
				p.WriteString("Should take " + target + " [carryable].\n")
			} else {
				p.WriteString("Should not take " + target + " [not carryable].\n")
			}
		} else {
			p.WriteString(target + " not seen.\n")
		}
	} else {
		p.WriteString("Take objects by typing 'take [object name]'.\n")
	}
}

func (p *Player) ReadLoop(playerRemoveChan chan *Player) {
	rawBuf := make([]byte, 1024)
	for {
		n, err := p.sock.Read(rawBuf)
		if err == nil {
			strCommand := string(rawBuf[:n])
			p.commandBuf <- strings.TrimRight(strCommand,"\n\r")
		} else if err == io.EOF {
			fmt.Println(p.name, "Disconnected")
			playerRemoveChan <- p
			return
		}
	}
}

func (p *Player) StimuliLoop() {
	for {
		nextStimulus := <- p.stimuli
		if p.DoesPerceive(nextStimulus) {
			p.WriteString(nextStimulus.Description(p))
		}
		fmt.Println(p.name,"receiving stimulus",nextStimulus.StimType())
	}
}

func (p *Player) HeartbeatLoop() {
	for {
		p.WriteString("Heartbeat\n")
		time.Sleep(5*time.Second)
	}
}

func (p *Player) WriteString(str string) {
	p.sock.Write([]byte(str))
}

func Divider() string { 
	return "\n-----------------------------------------------------------\n"
}

func (r *Room) Describe(toPlayer *Player) string {
	roomText := r.text
	objectsText := r.DescribeObjects(toPlayer)
	playersText := r.DescribePlayers(toPlayer)
	
	return roomText + Divider() + objectsText + Divider() + playersText
}

func (r *Room) DescribeObjects(toPlayer *Player) string {
	objTextBuf := "Sitting here is/are:\n"
	for _,obj := range r.physObjects {
		if obj.Visible() {
			objTextBuf += obj.Description()
			objTextBuf += "\n"
		}
	}
	return objTextBuf
}

func (r *Room) DescribePlayers(toPlayer *Player) string {
	objTextBuf := "Other people present:\n"
	for _,player := range r.players {
		if player.id != toPlayer.id {
			objTextBuf += player.name
			objTextBuf += "\n"
		}
	}
	return objTextBuf
}

func (s PlayerEnterStimulus) StimType() string { return "enter" }
func (s PlayerEnterStimulus) Description(p Perceiver) string {
	return s.player.name + " has entered the room.\n"
}

func (s PlayerLeaveStimulus) StimType() string { return "exit" }
func (s PlayerLeaveStimulus) Description(p Perceiver) string {
	return s.player.name + " has left the room.\n"
}

func (s PlayerSayStimulus) StimType() string { return "say" }
func (s PlayerSayStimulus) Description(p Perceiver) string {
	playerReceiver, ok := p.(*Player)
	if ok && s.player.id == playerReceiver.id {
		return "You say \"" + s.text + "\"\n"
	} 
	return s.player.name + " said " + "\"" + s.text + "\".\n"
}

func (s PlayerPickupStimulus) StimType() string { return "take" }
func (s PlayerPickupStimulus) Description(p Perceiver) string {
	playerReceiver, ok := p.(*Player)
	if ok && s.player.id == playerReceiver.id {
		return "You took " + s.obj.Description() + "\n"
	}
	return s.player.name + " took " + s.obj.Description() + ".\n"
}

func (p Player) DoesPerceive(s Stimulus) bool {
	switch s.(type) {
	case PlayerEnterStimulus: return p.DoesPerceiveEnter(s.(PlayerEnterStimulus))
        case PlayerLeaveStimulus: return p.DoesPerceiveExit(s.(PlayerLeaveStimulus))
	case PlayerSayStimulus: return true
	case PlayerPickupStimulus: return true
	}
	return false
}

// Would love to do away with this hack
func DemoteToPhysObjList(ps []Player) []PhysicalObject {
	physObjs := make([]PhysicalObject, len(ps))
	for idx, p := range(ps) { physObjs[idx] = p }
	return physObjs
}

func (p Player) PerceiveList() PerceiveMap {
	// Right now, perceive people in the room, objects in the room,
	// and objects in the player's inventory
	var targetList []PhysicalObject
	physObjects := make(PerceiveMap)
	room := roomList[p.room]
	people := room.players
	roomObjects := room.physObjects
	invObjects := p.inventory
	targetList = append(DemoteToPhysObjList(people), roomObjects...)
	targetList = append(targetList, invObjects...)

	for _,target := range(targetList) {
		if target.Visible() {
			for _,handle := range(target.TextHandles()) {
				physObjects[handle] = target
			}
		}
	}

	return physObjects
}

func (p Player) DoesPerceiveEnter(s PlayerEnterStimulus) bool {
	return !(s.player.id == p.id)
}

func (p Player) DoesPerceiveExit(s PlayerLeaveStimulus) bool {
	return !(s.player.id == p.id)
}

func AcceptConnAsPlayer(conn net.Conn, idSource func() int) *Player {
	// Make distinct unique names randomly
	colors := []string{"Red", "Blue", "Yellow"}
	animals := []string{"Pony", "Fox", "Jackal"}
	color := colors[rand.Intn(3)]
	animal := animals[rand.Intn(3)]
	p := new(Player)
	p.id = idSource()
	p.name = (color + animal)
	p.sock = conn
	p.commandBuf = make(chan string, 10)
	p.stimuli = make(chan Stimulus, 5)
	p.room = -1
	return p
}