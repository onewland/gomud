package mud

import ("net"
	"strings"
	"fmt"
	"io")

type Player struct {
	Talker
	Perceiver
	PhysicalObject
	id int
	room *Room
	name string
	sock net.Conn
	inventory []PhysicalObject
	universe *Universe
	commandBuf chan string
	stimuli chan Stimulus
	quitting chan bool
	commandDone chan bool
}

var colorMap map[string]string

func (p Player) ID() int { return p.id }
func (p Player) Name() string { return p.name }
func (p Player) StimuliChannel() chan Stimulus { return p.stimuli }

func init() {
	colorMap = make(map[string]string)
	colorMap["&black;"] = "\x1b[30m"
	colorMap["&red;"] = "\x1b[31m"
	colorMap["&green;"] = "\x1b[32m"
	colorMap["&yellow;"] = "\x1b[33m"
	colorMap["&blue;"] = "\x1b[34m"
	colorMap["&magenta;"] = "\x1b[35m"
	colorMap["&cyan;"] = "\x1b[36m"
	colorMap["&white;"] = "\x1b[37m"
	colorMap["&;"] = "\x1b[0m"

	GlobalCommands["who"] = Who
	GlobalCommands["look"] = Look
	GlobalCommands["say"] = Say
	GlobalCommands["take"] = Take
	GlobalCommands["go"] = GoExit
	GlobalCommands["inv"] = Inv
	GlobalCommands["quit"] = Quit
	GlobalCommands["make"] = Make
	
}

func (p Player) Room() *Room {
	return p.room
}

func RemovePlayerFromRoom(r *Room, p *Player) {
	delete(r.players, p.id)
	r.RemovePerceiver(p)
}

func PlacePlayerInRoom(r *Room, p *Player) {
	oldRoom := p.room
	if oldRoom != nil {
		oldRoom.stimuliBroadcast <- 
			PlayerLeaveStimulus{player: p}
		RemovePlayerFromRoom(oldRoom, p)
	}
	
	p.room = r
	r.stimuliBroadcast <- PlayerEnterStimulus{player: p}
	r.AddPerceiver(p)
	r.players[p.id] = *p
}

func (p Player) Visible() bool { return true }
func (p Player) Description() string { return "A person: " + p.name }
func (p Player) Carryable() bool { return false }
func (p Player) TextHandles() []string { 
	return []string{ strings.ToLower(p.name) } 
}

func (p *Player) TakeObject(o *PhysicalObject, r *Room) bool {
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
			if c, ok := GlobalCommands[nextCommandRoot]; ok {
				c(p, nextCommandArgs)
			}
		}
		p.WriteString("> ")
		p.commandDone <- true
	}
}

func Look(p *Player, args []string) {
	room := p.room
	if len(args) > 1 {
		fmt.Println("Too many args")
		p.WriteString("Too many args")
	} else {
		p.WriteString(room.Describe(p) + "\n")
	}
}

func Who(p *Player, args []string) {
	gotOne := false
	for id, pOther := range p.universe.Players {
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

func Say(p *Player, args []string) {
	room := p.room
	sayStim := TalkerSayStimulus{talker: p, text: strings.Join(args," ")}
	room.stimuliBroadcast <- sayStim
}

func Take(p *Player, args []string) {
	room := p.room
	if len(args) > 0 {
		target := strings.ToLower(args[0])
		room.interactionQueue <-
			PlayerTakeAction{ player: p, userTargetIdent: target }
	} else {
		p.WriteString("Take objects by typing 'take [object name]'.\n")
	}
}

func Inv(p *Player, args []string) {
	p.WriteString(Divider())
	for _, obj := range p.inventory {
		if obj != nil {
			p.WriteString(obj.Description())
		}
	}
	p.WriteString(Divider())
}

func GoExit(p *Player, args []string) {
	room := p.room
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

func Quit(p *Player, args[] string) {
	p.quitting <- true
}

func Make(p *Player, args[] string) {
	fmt.Println("[WARNING] Make command should not be in production")
	p.universe.Maker(p.universe, p, args)
}

func (p *Player) ReadLoop(playerRemoveChan chan *Player) {
	rawBuf := make([]byte, 1024)
	defer p.sock.Close()

	for ; ; <- p.commandDone {
		select {
		case <-p.quitting:
			playerRemoveChan <- p
			return
		default:
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
}

func (p *Player) HandleStimulus(s Stimulus) {
	p.WriteString(s.Description(p))
	fmt.Println(p.name,"receiving stimulus",s.StimType())
}

func (p *Player) WriteString(str string) {
	str_acc := str
	for easyCode, termCode := range colorMap {
		str_acc = strings.Replace(str_acc, easyCode, termCode, -1)
	}
	p.sock.Write([]byte(str_acc))
}

func (p Player) DoesPerceive(s Stimulus) bool {
	switch s.(type) {
	case PlayerEnterStimulus: 
		return p.DoesPerceiveEnter(s.(PlayerEnterStimulus))
	case PlayerLeaveStimulus: 
		return p.DoesPerceiveExit(s.(PlayerLeaveStimulus))
	case TalkerSayStimulus: return true
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

func (p Player) PerceiveList() PerceiveMap {
	// Right now, perceive people in the room, objects in the room,
	// and objects in the player's inventory
	var targetList []PhysicalObject
	physObjects := make(PerceiveMap)
	room := p.room
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