package mud

import ("strings"
	"fmt")

type PerceiveTest func(p Player, s Stimulus) bool
var PlayerPerceptions map[string]PerceiveTest

type Player struct {
	Talker
	Perceiver
	PhysicalObject
	Conn *UserConnection
	id int
	room *Room
	name string
	inventory []PhysicalObject
	Universe *Universe
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
	GlobalCommands["who"] = Who
	GlobalCommands["look"] = Look
	GlobalCommands["say"] = Say
	GlobalCommands["take"] = Take
	GlobalCommands["drop"] = Drop
	GlobalCommands["go"] = GoExit
	GlobalCommands["inv"] = Inv
	GlobalCommands["quit"] = Quit
	GlobalCommands["make"] = Make
	
	PlayerPerceptions = make(map[string]PerceiveTest)
	PlayerPerceptions["enter"] = DoesPerceiveEnter
	PlayerPerceptions["exit"] = DoesPerceiveExit
	PlayerPerceptions["say"] = DoesPerceiveSay
	PlayerPerceptions["take"] = DoesPerceiveTake
	PlayerPerceptions["drop"] = DoesPerceiveDrop
}

func (p Player) Room() *Room {
	return p.room
}

func RemovePlayerFromRoom(r *Room, p *Player) {
	delete(r.players, p.id)
	r.RemoveChild(p)
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
			(*o).SetRoom(nil)
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

func (p *Player) DropObject(o *PhysicalObject, r *Room) bool {
	Log("Dropping", o, "to", r)
	for idx, slot := range(p.inventory) {
		if(slot == *o) {
			r.AddChild(*o)
			p.inventory[idx] = nil
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
			} else {
				p.WriteString("Command '" + nextCommandRoot + "' not recognized.\n")
			}
		}
		p.WriteString("> ")
		p.commandDone <- true
	}
}

func Look(p *Player, args []string) {
	room := p.room
	if len(args) > 1 {
		Log("Too many args")
		p.WriteString("Too many args")
	} else {
		p.WriteString(room.Describe(p) + "\n")
	}
}

func Who(p *Player, args []string) {
	gotOne := false
	for id, pOther := range p.Universe.Players {
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

func Drop(p *Player, args []string) {
	room := p.room
	if len(args) > 0 {
		target := strings.ToLower(args[0])
		room.interactionQueue <-
			PlayerDropAction{ player: p, userTargetIdent: target }
	} else {
		p.WriteString("Drop objects by typing 'drop [object name]'.\n")
	}
}

func Inv(p *Player, args []string) {
	p.WriteString(Divider())
	p.WriteString("Inventory: \n")
	for _, obj := range p.inventory {
		if obj != nil {
			p.WriteString(obj.Description())
			p.WriteString("\n")
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

	room.WithExit(args[0], func(foundExit *RoomExitInfo) {
		PlacePlayerInRoom(foundExit.OtherSide(), p)
		Look(p, []string{})
	}, func() {
		p.WriteString("No visible exit " + args[0] + ".\n")
	})
}

func Quit(p *Player, args[] string) {
	p.quitting <- true
}

func Make(p *Player, args[] string) {
	Log("[WARNING] Make command should not be in production")
	p.Universe.Maker(p.Universe, p, args)
}

func (p *Player) ReadLoop(playerRemoveChan chan *Player) {
	for ; ; <- p.commandDone {
		select {
		case <-p.quitting:
			Log("quitting in ReadLoop")
			playerRemoveChan <- p
			p.WriteString("Goodbye!")
			p.Conn.Close()
			return
		case c := <- p.Conn.FromUser:
			p.commandBuf <- c
		}
	}
}

func (p *Player) HandleStimulus(s Stimulus) {
	p.WriteString(s.Description(p))
	Log(p.name,"receiving stimulus",s.StimType())
}

func (p *Player) WriteString(str string) { p.Conn.Write(str) }

func (p Player) DoesPerceive(s Stimulus) bool {
	perceptTest := PlayerPerceptions[s.StimType()]
	return perceptTest(p, s)
}

func DoesPerceiveEnter(p Player, s Stimulus) bool {
	sEnter, ok := s.(PlayerEnterStimulus)
	if !ok { panic("Bad input to DoesPerceiveEnter") }
	return !(sEnter.player.id == p.id)
}

func DoesPerceiveExit(p Player, s Stimulus) bool {
	sExit, ok := s.(PlayerLeaveStimulus)
	if !ok { panic("Bad input to DoesPerceiveExit") }
	return !(sExit.player.id == p.id)
}

func DoesPerceiveSay(p Player, s Stimulus) bool { return true }
func DoesPerceiveTake(p Player, s Stimulus) bool { return true }
func DoesPerceiveDrop(p Player, s Stimulus) bool { return true }

func (p Player) PerceiveList(context PerceiveContext) PerceiveMap {
	// Right now, perceive people in the room, objects in the room,
	// and objects in the player's inventory
	var targetList []PhysicalObject
	physObjects := make(PerceiveMap)
	room := p.room
	people := room.players
	roomObjects := room.physObjects
	invObjects := p.inventory

	targetList = []PhysicalObject{}

	if(context == LookContext) {
		targetList = append(targetList, PlayersAsPhysObjSlice(people)...)
	}
	if(context == TakeContext || context == LookContext) {
		targetList = append(targetList, roomObjects...)
	}
	if(context == InvContext || context == LookContext) {
		targetList = append(targetList, invObjects...)
	}

	for _,target := range(targetList) {
		Log(target)
		if target != nil && target.Visible() {
			for _,handle := range(target.TextHandles()) {
				physObjects[handle] = target
			}
		}
	}

	return physObjects
}