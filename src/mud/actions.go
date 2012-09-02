package mud

type PlayerTakeAction struct {
	InterObjectAction
	player *Player
	target PhysicalObject
	userTargetIdent string
}

func noSpaceMsg(name string) string {
	return "No space in your inventory for " + name + ".\n"
}
func noCarryMsg(name string) string {
	return name + " cannot be carried.\n"
}

func (p PlayerTakeAction) Targets() []PhysicalObject {
	targets := make([]PhysicalObject, 1)
	targets[0] = p.target
	return targets
}
func (p PlayerTakeAction) Source() PhysicalObject { return p.player }
func (p PlayerTakeAction) Exec() {
	player := p.player
	room := player.room
	if target, ok := player.PerceiveList(TakeContext)[p.userTargetIdent]; ok {
		stim := PlayerPickupStimulus{player: player, obj: target}
		if target.Carryable() {
			if player.TakeObject(&target, room) {
				room.stimuliBroadcast <- stim
			} else {
				player.WriteString(noSpaceMsg(p.userTargetIdent))
			}
		} else {
			player.WriteString(noCarryMsg(p.userTargetIdent))
		}
	} else {
		player.WriteString(p.userTargetIdent + " not seen.\n")
	}
}

type PlayerDropAction struct {
	InterObjectAction
	player *Player
	target PhysicalObject
	userTargetIdent string
}

func (p PlayerDropAction) Targets() []PhysicalObject {
	targets := make([]PhysicalObject, 1)
	targets[0] = p.target
	return targets
}
func (p PlayerDropAction) Source() PhysicalObject { return p.player }
func (p PlayerDropAction) Exec() {
	player := p.player
	room := player.room
	if target, ok := player.PerceiveList(InvContext)[p.userTargetIdent]; ok {
		stim := PlayerDropStimulus{player: player, obj: target}
		
		if player.DropObject(&target, room) {
			room.stimuliBroadcast <- stim
		} else {
			player.WriteString("Object cannot be dropped.\n")
		}
	} else {
		player.WriteString(p.userTargetIdent + " not in your inventory.\n")
	}
}