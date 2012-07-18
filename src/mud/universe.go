package mud

import ("net" 
	"fmt" 
	"math/rand")

type Universe struct {
	PlayerList map[int]*Player
	RoomList map[RoomID]*Room
}

func NewBasicUniverse() *Universe {
	u := new(Universe)
	u.PlayerList = make(map[int]*Player)
	u.RoomList = make(map[RoomID]*Room)
	return u
}

func (u *Universe) AcceptConnAsPlayer(conn net.Conn, idSource func() int) *Player {
	// Make distinct unique names randomly
	colors := []string{"Red", "Blue", "Yellow"}
	animals := []string{"Pony", "Fox", "Jackal"}
	color := colors[rand.Intn(3)]
	animal := animals[rand.Intn(3)]
	p := new(Player)
	p.id = idSource()
	p.name = (color + animal)
	p.sock = conn
	p.quitting = make(chan bool, 1)
	p.commandBuf = make(chan string, 10)
	p.commandDone = make(chan bool)
	p.stimuli = make(chan Stimulus, 5)
	p.inventory = make([]PhysicalObject, 10)
	p.room = -1
	p.universe = u
	PlayerList[p.id] = p
	fmt.Println(p.name, "joined, ID =",p.id)
	fmt.Println(len(PlayerList), "player[s] online.")
	return p
}