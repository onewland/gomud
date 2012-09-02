package mud

import ("strconv"; "strings")

func init() {
	PersistentKeys["room"] = []string{ "id", "text", "persisters" }
	PersistentKeys["roomConnect"] = []string{ 
		"id", "aExitName", 
		"bExitName", "roomAId", "roomBId" }

	containerHelper := new(FlexObjHandlerPair)
	containerHelper.Add = func(fc *FlexContainer, o interface{}) {
		ifIsPhysical(o, func(p PhysicalObject) {
			r := fc.Meta["Room"].(*Room)
			p.SetRoom(r)
		})
	}
	containerHelper.Remove = func(*FlexContainer, interface{}) {}
	FlexObjHandlers["RoomPhysicalObjects"] = *containerHelper
}

type RoomID int
type RoomSide int

type RoomConnCreator func() *SimpleRoomConnection
type RoomConnector func(a *Room, b *Room) *SimpleRoomConnection

const (
	SideA RoomSide = iota
	SideB
)

type RoomExitInfo struct {
	exit RoomConnection
	exitSide RoomSide
}

type Room struct {
	Persister
	CommandSource
	id int
	text string
	players map[int]*Player
	children *FlexContainer
	exits []RoomExitInfo
	stimuliBroadcast chan Stimulus
	interactionQueue chan InterObjectAction
	universe *Universe
}

type RoomConnection interface {
	RoomA() *Room
	RoomB() *Room
	AExitName() string
	BExitName() string
}

type SimpleRoomConnection struct {
	Persister
	RoomConnection
	id int
	roomA, roomB *Room
	aExitName, bExitName string
}

func (rc SimpleRoomConnection) RoomA() *Room { return rc.roomA }
func (rc SimpleRoomConnection) RoomB() *Room { return rc.roomB }
func (rc SimpleRoomConnection) AExitName() string { return rc.aExitName }
func (rc SimpleRoomConnection) BExitName() string { return rc.bExitName }

func (rc SimpleRoomConnection) PersistentValues() map[string]interface{} {
	vals := make(map[string]interface{})
	if(rc.id > 0) {
		vals["id"] = strconv.Itoa(rc.id)
	}
	vals["roomAId"] = strconv.Itoa(rc.roomA.id)
	vals["roomBId"] = strconv.Itoa(rc.roomB.id)
	vals["aExitName"] = rc.aExitName
	vals["bExitName"] = rc.bExitName
	return vals
}

func (rc *SimpleRoomConnection) Save() string {
	universe := rc.roomA.universe
	outID := universe.Store.SaveStructure("roomConnect",rc.PersistentValues())
	if(rc.id == 0) {
		rc.id, _ = strconv.Atoi(outID)
		universe.Store.AddToGlobalSet("roomConnects", outID)
	}
	return outID
}

func LoadRoomConn(universe *Universe, id int) *SimpleRoomConnection {
	//Log("Entering LoadRoomConn")
	vals := universe.Store.LoadStructure(PersistentKeys["roomConnect"],
		FieldJoin(":","roomConnect",strconv.Itoa(id)))
	roomAIdStr, _ := vals["roomAId"].(string)
	roomBIdStr, _ := vals["roomBId"].(string)
	aExitName,_ := vals["aExitName"].(string)
	bExitName,_ := vals["bExitName"].(string)
	cc := SimpleRoomConnectCreator(aExitName,bExitName)
	conn := ConnectWithConnCreator(cc)
	roomAId,_ := strconv.Atoi(roomAIdStr)
	roomBId,_ := strconv.Atoi(roomBIdStr)
	rc := conn(universe.Rooms[roomAId],universe.Rooms[roomBId])
	rc.id = id
	universe.Add(rc)
	return rc
}

func SimpleRoomConnectCreator(a string, b string) RoomConnCreator {
	return func() *SimpleRoomConnection {
		conn := new(SimpleRoomConnection)
		conn.aExitName, conn.bExitName = a, b
		return conn
	}
}

var EastWestRoomConnection = SimpleRoomConnectCreator("east","west")
var NorthSouthRoomConnection = SimpleRoomConnectCreator("north","south")
var UpDownRoomConnection = SimpleRoomConnectCreator("up","down")

func ConnectWithConnCreator(exitGen RoomConnCreator) RoomConnector {
	return func(a *Room, b *Room) *SimpleRoomConnection {
		roomConn := exitGen()
		roomConn.roomA = a
		roomConn.roomB = b
		reiA := RoomExitInfo{exitSide: SideA, exit: roomConn}
		reiB := RoomExitInfo{exitSide: SideB, exit: roomConn}
		a.exits = append(a.exits, reiA)
		b.exits = append(b.exits, reiB)
		return roomConn
	}
}

var ConnectEastWest = ConnectWithConnCreator(EastWestRoomConnection)
var ConnectNorthSouth = ConnectWithConnCreator(NorthSouthRoomConnection)
var ConnectUpDown = ConnectWithConnCreator(UpDownRoomConnection)

func (r *Room) ActionQueue() {
	for {
		action := <- r.interactionQueue
		action.Exec()
	}
}

func (r *Room) Perceivers() []Perceiver {
	return castAsPerceivers(r.children.AllObjects["Perceivers"])
}

func (r *Room) FanOutBroadcasts() {
	for {
		broadcast := <- r.stimuliBroadcast
		for _,p := range r.Perceivers() { 
			//Log("fanning broadcast to ",p)
			p.StimuliChannel() <- broadcast 
		}
	}
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

func (r *Room) Describe(toPlayer *Player) string {
	roomText := r.text
	objectsText := r.DescribeObjects(toPlayer)
	playersText := r.DescribePlayers(toPlayer)
	exitsText := "Exits: " + r.ExitNames()
	
	return roomText + Divider() + 
		objectsText + Divider() + 
		playersText + exitsText
}

func (r *Room) DescribeObjects(toPlayer *Player) string {
	objTextBuf := "Sitting here is/are:\n"
	for _,obj := range r.PhysicalObjects() {
		if obj != nil && obj.Visible() {
			objTextBuf += obj.Description()
			objTextBuf += "\n"
		}
	}
	return objTextBuf
}

func (r *Room) DescribePlayers(toPlayer *Player) string {
	if len(r.players) > 1 {
		objTextBuf := "Other people present:\n"
		for _,player := range r.players {
			if player.id != toPlayer.id {
				objTextBuf += player.name
				objTextBuf += "\n"
			}
		}
		return objTextBuf
	}
	return ""
}

func (r *Room) AddChild(o interface{}) {
	r.children.Add(o)
}

func (r *Room) RemoveChild(o interface{}) {
	r.children.Remove(o)
}

func (r *Room) Broadcast(s Stimulus) {
	r.stimuliBroadcast <- s
}

func (r Room) PersistentValues() map[string]interface{} {
	vals := make(map[string]interface{})
	if(r.id > 0) {
		vals["id"] = strconv.Itoa(r.id)
	}
	vals["text"] = r.text
	vals["persisters"] = castAsPersistents(
		r.children.AllObjects["Persistents"])
	return vals
}

func (r *Room) Save() string {
	outID := r.universe.Store.SaveStructure("room",r.PersistentValues())
	if(r.id == 0) {
		r.id, _ = strconv.Atoi(outID)
	}
	r.universe.Store.AddToGlobalSet("rooms", outID)
	return outID
}

func LoadRoom(universe *Universe, id int) *Room {
	vals := universe.Store.LoadStructure(PersistentKeys["room"],
		FieldJoin(":","room",strconv.Itoa(id)))
	Log("LoadRoom vals",vals)
	if textStr, ok := vals["text"].(string); ok {
		r := NewRoom(universe, id, textStr)
		if persisterIds, ok := vals["persisters"].([]string); ok {
			for _,pid := range(persisterIds) {
				p := LoadArbitrary(universe, pid)
				r.AddChild(p)
			}
		}
		return r
	}
	return nil
}

func NewRoom(universe *Universe, rid int, rtext string) *Room {
	r := Room{id: rid, text: rtext, universe: universe}
	r.stimuliBroadcast = make(chan Stimulus, 10)
	r.interactionQueue = make(chan InterObjectAction, 10)
	r.players = make(map[int]*Player)
	r.exits = []RoomExitInfo{}
	r.children = NewFlexContainer(
		"PhysicalObjects",
		"Persistents",
		"RoomPhysicalObjects",
		"Perceivers",
		"CommandSources")
	r.children.Meta["Room"] = &r
	universe.Rooms[r.id] = &r
	universe.Add(&r)

	go r.FanOutBroadcasts()
	go r.ActionQueue()

	return &r
}

type FoundExit func(rei *RoomExitInfo)

func (r *Room) WithExit(name string, found FoundExit, notFound func()) {
	for _,exit := range(r.exits) {
		if name == exit.Name() {
			found(&exit)
			return
		}
	}
	notFound()
}

func (r *Room) SetText(text string) { r.text = text }
func (r *Room) Text() string { return r.text }

type PhysObjReceiver func(p *PhysicalObject)

func (r *Room) WithPhysObjects(handler PhysObjReceiver) {
	for _,p := range(r.PhysicalObjects()) { handler(&p) }
}

func (r *Room) ExitNames() string {
	exitNames := make([]string, len(r.exits))
	for i,exit := range(r.exits) {
		exitNames[i] = exit.Name()
	}
	return strings.Join(exitNames, ", ")
}

func (r *Room) Actions() chan InterObjectAction { return r.interactionQueue }

func (r *Room) CommandSources() []CommandSource {
	return castCmdSources(r.children.AllObjects["CommandSources"])
}

func (r *Room) Commands() map[string]Command {
	localCommands := make(map[string]Command)
	for _, source := range(r.CommandSources()) {
		for commandName, command := range(source.Commands()) {
			localCommands[commandName] = command
		}
	}
	return localCommands
}

func castCmdSources(o []interface{}) []CommandSource {
	pos := make([]CommandSource, len(o))
	for i, x := range(o) { po := x.(CommandSource); pos[i] = po }
	return pos
}

func castPhysicalObjects(o []interface{}) []PhysicalObject {
	pos := make([]PhysicalObject, len(o))
	for i, x := range(o) { po := x.(PhysicalObject); pos[i] = po }
	return pos
}

func castAsPersistents(o []interface{}) []Persister {
	pos := make([]Persister, len(o))
	for i, x := range(o) { po := x.(Persister); pos[i] = po }
	return pos
}

func castAsPerceivers(o []interface{}) []Perceiver {
	pos := make([]Perceiver, len(o))
	for i, x := range(o) { po := x.(Perceiver); pos[i] = po }
	return pos
}

func (r *Room) PhysicalObjects() []PhysicalObject {
	return castPhysicalObjects(r.children.AllObjects["PhysicalObjects"])
}