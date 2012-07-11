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
 * 5. Heartbeat
 *
 * Initially not supporting:
 * - Persistence of state
 * - Combat
 * - NPCs
 */

type InterObjectAction interface {
	Targets() []PhysicalObject
	Source() PhysicalObject
	Exec()
}

type PlayerTakeAction struct {
	InterObjectAction
	player *Player
	target PhysicalObject
	userTargetIdent string
}

func (p PlayerTakeAction) Targets() []PhysicalObject {
	targets := make([]PhysicalObject, 1)
	targets[0] = p.target
	return targets
}
func (p PlayerTakeAction) Source() PhysicalObject { return p.player }
func (p PlayerTakeAction) Exec() {
	fmt.Println("exec take",p.target,p.Source())
	player := p.player
	room := roomList[player.room]
	if target, ok := player.PerceiveList()[p.userTargetIdent]; ok {
		if target.Carryable() {
			if player.PlaceObjectInInventoryFromRoom(&target, room) {
				room.stimuliBroadcast <- PlayerPickupStimulus{player: player, obj: target}
			} else {
				player.WriteString("No space in your inventory for " + p.userTargetIdent + ".\n")
			}
		} else {
			player.WriteString("Should not take " + p.userTargetIdent + " [not carryable].\n")
		}
	} else {
		player.WriteString(p.userTargetIdent + " not seen.\n")
	}
}

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
	from string
}

type PlayerLeaveStimulus struct {
	Stimulus
	player *Player
	to string
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
type RoomSide int
const (
	SideA RoomSide = iota
	SideB
)

type RoomExitInfo struct {
	exit RoomConnection
	exitSide RoomSide
}

type Room struct {
	id RoomID
	text string
	players map[int]Player
	physObjects []PhysicalObject
	exits []RoomExitInfo
	stimuliBroadcast chan Stimulus
	interactionQueue chan InterObjectAction
}

type RoomConnection interface {
	RoomA() *Room
	RoomB() *Room
	AExitName() string
	BExitName() string
}

type EastWestRoomConnection struct {
	RoomConnection
	roomA *Room
	roomB *Room
}

func (rc EastWestRoomConnection) RoomA() *Room { return rc.roomA }
func (rc EastWestRoomConnection) RoomB() *Room { return rc.roomB }
func (rc EastWestRoomConnection) AExitName() string { return "east" }
func (rc EastWestRoomConnection) BExitName() string { return "west" }

type PhysicalObject interface {
	Visible() bool
	Description() string
	Carryable() bool
	TextHandles() []string
	TakeReady() chan bool
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

// 'a' side will be west of 'b' side
func ConnectEastWest(a *Room, b *Room) *EastWestRoomConnection {
	roomConn := EastWestRoomConnection{roomA: a, roomB: b}
	reiA := RoomExitInfo{exitSide: SideA, exit: roomConn}
	reiB := RoomExitInfo{exitSide: SideB, exit: roomConn}
	a.exits = append(a.exits, reiA)
	b.exits = append(b.exits, reiB)
	return &roomConn
}

func MakeStupidRoom() *Room {
	room := Room{id: 1, text: "You are in a bedroom." }
	room2 := Room{id: 2, text: "You are in a bathroom." }

	room.stimuliBroadcast = make(chan Stimulus, 10)
	room2.stimuliBroadcast = make(chan Stimulus, 10)
	room.interactionQueue = make(chan InterObjectAction, 10)
	room2.interactionQueue = make(chan InterObjectAction, 10)
	room.players = make(map[int]Player)
	room2.players = make(map[int]Player)

	ConnectEastWest(&room, &room2)
	theBall := Ball{}
	room.physObjects = []PhysicalObject {theBall}

	roomList[room.id] = &room
	roomList[room2.id] = &room2

	go room.FanOutBroadcasts()
	go room2.FanOutBroadcasts()
	go room.ActionQueue()
	go room2.ActionQueue()

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

	go HeartbeatLoop()

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
		delete(oldRoom.players, p.id)
	}
	
	p.room = r.id
	r.stimuliBroadcast <- PlayerEnterStimulus{player: p}
	r.players[p.id] = *p
}

func (p *Player) PlaceObjectInInventoryFromRoom(o *PhysicalObject, r *Room) bool {
	for idx, slot := range(p.inventory) {
		if(slot == nil) {
			p.inventory[idx] = *o
			for idx, obj := range(r.physObjects) {
				if(*o == obj) {
					r.physObjects[idx] = nil
					break
				}
			}
			return true
		}
	}
	return false
}

func UniqueIDGen() func() int {
	x, xchan := 0, make(chan int)
	go func() { for ; ; x += 1 { xchan <- x } }()

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
	re, _ := regexp.Compile(`(\S+)|(['"][^'"]+['"])`)
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
			} else if nextCommandRoot == "go" {
				p.GoExit(nextCommandArgs)
			} else if nextCommandRoot == "inv" {
				p.Inv(nextCommandArgs)
			}
		}
		p.WriteString("> ")
	}
}

func (r *Room) ActionQueue() {
	for {
		action := <- r.interactionQueue
		action.Exec()
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
		room.interactionQueue <-
			PlayerTakeAction{ player: p, userTargetIdent: target }
	} else {
		p.WriteString("Take objects by typing 'take [object name]'.\n")
	}
}

func (p *Player) Inv(args []string) {
	p.WriteString(Divider())
	for _, obj := range p.inventory {
		if obj != nil {
			p.WriteString(obj.Description())
		}
	}
	p.WriteString(Divider())
}

func (r *RoomExitInfo) Name() string {
	if(r.exitSide == SideA) {
		return r.exit.AExitName()
	} else {
		return r.exit.BExitName()
	}
	return ""
}

func (r *RoomExitInfo) OtherSide() *Room {
	if(r.exitSide == SideA) {
		return r.exit.RoomB()
	} else {
		return r.exit.RoomA()
	}
	return nil
}

func (p *Player) GoExit(args []string) {
	room := roomList[p.room]
	if(len(args) < 1) {
		p.WriteString("Go usage: go [exit name]. Ex. go north")
		return 
	}
	var foundExit *RoomExitInfo
	for _,exit := range(room.exits) {
		if args[0] == exit.Name() {
			foundExit = &exit
			break
		}
	}
	if foundExit != nil {
		PlacePlayerInRoom(foundExit.OtherSide(), p)
		p.WriteString("Should go through exit " + foundExit.Name())
	} else {
		p.WriteString("No visible exit " + args[0] + ".\n")
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

func HeartbeatLoop() {
	//recurScheduled := make(map[int]list.List)

	for n:=0 ; ; n++ {
		//fmt.Println("Heartbeat", time.Now())
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
		if obj != nil && obj.Visible() {
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

func PlayersAsPhysObjSlice(ps map[int]Player) []PhysicalObject {
	physObjs := make([]PhysicalObject, len(ps))
	n := 0
	for _, p := range(ps) { 
		physObjs[n] = p 
		n++
	}
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
	targetList = append(PlayersAsPhysObjSlice(people), roomObjects...)
	targetList = append(targetList, invObjects...)

	for _,target := range(targetList) {
		fmt.Println(target)
		if target != nil && target.Visible() {
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
	p.inventory = make([]PhysicalObject, 10)
	p.room = -1
	return p
}
