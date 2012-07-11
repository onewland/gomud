package mud

var RoomList map[RoomID]*Room;

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

// 'a' side will be west of 'b' side
func ConnectEastWest(a *Room, b *Room) *EastWestRoomConnection {
	roomConn := new(EastWestRoomConnection)
	roomConn.roomA = a
	roomConn.roomB = b
	reiA := RoomExitInfo{exitSide: SideA, exit: roomConn}
	reiB := RoomExitInfo{exitSide: SideB, exit: roomConn}
	a.exits = append(a.exits, reiA)
	b.exits = append(b.exits, reiB)
	return roomConn
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