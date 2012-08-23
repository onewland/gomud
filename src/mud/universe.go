package mud

import ("time"
        "redis")

type MakeHandler func (*Universe, *Player, []string)

type Universe struct {
	Players map[int]*Player
	Rooms map[int]*Room
	TimeListeners []TimeListener
	Persistents []Persister
	Maker MakeHandler
	Store *TinyDB
	dbConn redis.Client
}

func NewBasicUniverse() *Universe {
	u := new(Universe)
	u.Players = make(map[int]*Player)
	u.Rooms = make(map[int]*Room)
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

func (u *Universe) AddPersistent(p Persister) {
	u.Persistents = append(u.Persistents, p)
}


func (u *Universe) HeartbeatLoop(speedupFactor float64) {
	for n:=0 ; ; n++ {
		for _, l := range(u.TimeListeners) {
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
