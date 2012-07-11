package mud

var RoomList map[RoomID]*Room;

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

type SimpleRoomConnection struct {
	RoomConnection
	roomA, roomB *Room
	aExitName, bExitName string
}

func (rc SimpleRoomConnection) RoomA() *Room { return rc.roomA }
func (rc SimpleRoomConnection) RoomB() *Room { return rc.roomB }
func (rc SimpleRoomConnection) AExitName() string { return rc.aExitName }
func (rc SimpleRoomConnection) BExitName() string { return rc.bExitName }

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

func (r *Room) FanOutBroadcasts() {
	for {
		broadcast := <- r.stimuliBroadcast
		for _,p := range r.players { 
			p.stimuli <- broadcast 
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

func NewBasicRoom(rid RoomID, rtext string, physObjs []PhysicalObject) *Room {
	r := Room{id: rid, text: rtext}
	r.stimuliBroadcast = make(chan Stimulus, 10)
	r.interactionQueue = make(chan InterObjectAction, 10)
	r.players = make(map[int]Player)
	r.physObjects = physObjs
	r.exits = []RoomExitInfo{}
	RoomList[r.id] = &r

	return &r
}