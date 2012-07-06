package main

import ("fmt"
	"net"
	"math/rand"
	"time"
	"io"
	"strings"
	"sync"
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

type RoomID int

type Room struct {
	id RoomID
	text string
	players []Player
	physObjects []PhysicalObject
}

type PhysicalObject interface {
	Visible() bool
	Description() string
	Carryable() bool
}

type Ball struct { PhysicalObject }

func (b Ball) Visible() bool { return true }
func (b Ball) Description() string { return "A red ball." }
func (b Ball) Carryable() bool { return true }

type Player struct {
	id int
	room RoomID
	name string
	sock net.Conn
	commandBuf chan string
}

var playerList map[int]*Player
var roomList map[RoomID]*Room

func main() {
	rand.Seed(time.Now().Unix())
	listener, err := net.Listen("tcp", ":3000")
	playerRemoveChan := make(chan *Player)
	playerList = make(map[int]*Player)
	roomList = make(map[RoomID]*Room)
	idGen := UniqueIDGen()
	theBall := Ball{}
	theRoom := Room{id: 1, text:"You are in a bedroom."}
	theRoom.physObjects = []PhysicalObject {theBall}
	roomList[theRoom.id] = &theRoom

	if err == nil {
		go PlayerListManager(playerRemoveChan, playerList)
		defer listener.Close()

		fmt.Println("Listening on port 3000")
		for {
			conn, aerr := listener.Accept()
			if aerr == nil {
				newP := AcceptConnAsPlayer(conn, idGen)
				newP.room = theRoom.id
				theRoom.players = append(theRoom.players, *newP)
				playerList[newP.id] = newP
				fmt.Println(newP.name, "joined, ID =",newP.id)
				fmt.Println(len(playerList), "player[s] online.")

				//go newP.HeartbeatLoop()
				go newP.ReadLoop(playerRemoveChan)
				go newP.ExecCommandLoop()
			} else {
				fmt.Println("Error in accept")
			}
		}
	} else {
		fmt.Println("Error in listen")
	}
}

func UniqueIDGen() func() int {
	x := 0
	m := new(sync.Mutex)
	return func() int {
		m.Lock()
		x += 1
		return_x := x
		m.Unlock()
		return return_x
	}
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
			if nextCommandRoot == "who" { p.Who(nextCommandArgs) }
			if nextCommandRoot == "look" { p.Look(nextCommandArgs) }
		}
	}
}

func (p *Player) Look(args []string) {
	if len(args) > 1 {
		fmt.Println("Too many args")
		p.sock.Write([]byte("Too many args"))
	} else {
		p.sock.Write([]byte(roomList[p.room].Describe(p) + "\n"))
	}
}

func (p *Player) Who(args []string) {
	gotOne := false
	for id, pOther := range playerList {
		if id != p.id {
			str_who := fmt.Sprintf("[WHO] %s\n",pOther.name)
			p.sock.Write([]byte(str_who))
			gotOne = true
		}
	}

	if !gotOne {
		p.sock.Write([]byte("You are all alone in the world.\n"))
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

func (p *Player) HeartbeatLoop() {
	for {
		p.sock.Write([]byte("Heartbeat\n"))
		time.Sleep(5*time.Second)
	}
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

func AcceptConnAsPlayer(conn net.Conn, idSource func() int) *Player {
	// Make distinct unique names randomly
	colors := []string{"Red", "Blue", "Yellow"}
	animals := []string{"Pony", "Fox", "Jackal"}
	color := colors[rand.Intn(3)]
	animal := animals[rand.Intn(3)]
	p := new(Player)
	p.id = idSource()
	p.name = (color + " " + animal)
	p.sock = conn
	p.commandBuf = make(chan string, 10)
	return p
}