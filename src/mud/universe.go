package mud

import ("time"
        "redis")

type MakeHandler func (*Universe, *Player, []string)

type Universe struct {
	Players map[int]*Player
	Rooms map[int]*Room
	children *FlexContainer
	Maker MakeHandler
	Store *TinyDB
	dbConn redis.Client
}

func NewUniverse() *Universe {
	u := new(Universe)
	u.Players = make(map[int]*Player)
	u.Rooms = make(map[int]*Room)
	u.children = NewFlexContainer("Persistents", "TimeListeners")
	spec := redis.DefaultSpec().Db(3)
	client, err := redis.NewSynchClientWithSpec(spec)
	if(err != nil) {
		panic(err)
	} else {
		u.dbConn = client
		u.Store = NewTinyDB(client)
	}
	return u
}

func (u *Universe) TimeListeners() []TimeListener {
	return castTimeListeners(u.children.AllObjects["TimeListeners"])
}

func (u *Universe) Persistents() []Persister {
	return castAsPersistents(u.children.AllObjects["Persistents"])
}

func (u *Universe) PlayerFromUserConn(conn *UserConnection, idSource func() int) *Player {
	name := conn.Data["playerName"].(string)
	p := CreateOrLoadPlayer(u, name)
	p.Conn = conn
	u.Players[p.id] = p
	Log(p.name, "joined, ID =",p.id)
	Log(len(u.Players), "player[s] online.")

	conn.OnDisconnect = func() {
		Log(p.name, "Disconnecting")
		p.quitting <- true
		p.commandDone <- true
	}
	return p
}

func (u *Universe) ClearDB() {
	u.Store.Flush()
}

func (u *Universe) Add(o interface{}) {
	Log("universe add")
	u.children.Add(o)
}

func (u *Universe) HeartbeatLoop(speedupFactor float64) {
	for n:=0 ; ; n++ {
		for _, l := range(u.TimeListeners()) {
			l.Ping() <- n
		}
		time.Sleep(time.Duration(int(1000000/speedupFactor))*time.Nanosecond)
	}
}

func PlayerListManager(toRemove chan *Player, pList map[int]*Player) {
	for {
		pRemove := <- toRemove
		pRoom := pRemove.room
		RemovePlayerFromRoom(pRoom, pRemove)
		delete(pList, pRemove.id)
		Log("Removed", pRemove.name, "from player list")
	}
}

func castTimeListeners(o []interface{}) []TimeListener {
	pos := make([]TimeListener, len(o))
	for i, x := range(o) { po := x.(TimeListener); pos[i] = po }
	return pos
}