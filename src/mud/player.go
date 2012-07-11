package mud

import ("net"
	"math/rand"
	"strings"
	"fmt"
	"io")

var PlayerList map[int]*Player

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

func PlacePlayerInRoom(r *Room, p *Player) {
	oldRoomID := p.room
	if oldRoomID != -1 {
		oldRoom := RoomList[oldRoomID]
		oldRoom.stimuliBroadcast <- 
			PlayerLeaveStimulus{player: p}
		delete(oldRoom.players, p.id)
	}
	
	p.room = r.id
	r.stimuliBroadcast <- PlayerEnterStimulus{player: p}
	r.players[p.id] = *p
}

func (p Player) Visible() bool { return true }
func (p Player) Description() string { return "A person: " + p.name }
func (p Player) Carryable() bool { return false }
func (p Player) TextHandles() []string { return []string{ strings.ToLower(p.name) } }

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

func (p *Player) Look(args []string) {
	if len(args) > 1 {
		fmt.Println("Too many args")
		p.WriteString("Too many args")
	} else {
		p.WriteString(RoomList[p.room].Describe(p) + "\n")
	}
}

func (p *Player) Who(args []string) {
	gotOne := false
	for id, pOther := range PlayerList {
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
	room := RoomList[p.room]
	sayStim := PlayerSayStimulus{player: p, text: strings.Join(args," ")}
	room.stimuliBroadcast <- sayStim
}

func (p *Player) Take(args []string) {
	room := RoomList[p.room]
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

func (p *Player) GoExit(args []string) {
	room := RoomList[p.room]
	if(len(args) < 1) {
		p.WriteString("Go usage: go [exit name]. Ex. go north")
		return 
	}
	var foundExit *RoomExitInfo
	fmt.Println(room)
	fmt.Println(room.exits)
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

func (p *Player) WriteString(str string) {
	p.sock.Write([]byte(str))
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
	PlayerList[p.id] = p
	fmt.Println(p.name, "joined, ID =",p.id)
	fmt.Println(len(PlayerList), "player[s] online.")
	return p
}

func PlayerListManager(toRemove chan *Player, pList map[int]*Player) {
	for {
		pRemove := <- toRemove
		delete(pList, pRemove.id)
		fmt.Println("Removed", pRemove.name, "from player list")
	}
}

func (p Player) PerceiveList() PerceiveMap {
	// Right now, perceive people in the room, objects in the room,
	// and objects in the player's inventory
	var targetList []PhysicalObject
	physObjects := make(PerceiveMap)
	room := RoomList[p.room]
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