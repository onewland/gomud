package mud

type PlayerTakeAction struct {
	InterObjectAction
	player *Player
	target PhysicalObject
	userTargetIdent string
}


func NoSpaceMsg(name string) string {
	return "No space in your inventory for " + name + ".\n"
}
func NoCarryMsg(name string) string {
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
	universe := player.universe
	room := universe.Rooms[player.room]
	if target, ok := player.PerceiveList()[p.userTargetIdent]; ok {
		stim := PlayerPickupStimulus{player: player, obj: target}
		if target.Carryable() {
			if player.TakeObject(&target, room) {
				room.stimuliBroadcast <- stim
			} else {
				player.WriteString(NoSpaceMsg(p.userTargetIdent))
			}
		} else {
			player.WriteString(NoCarryMsg(p.userTargetIdent))
		}
	} else {
		player.WriteString(p.userTargetIdent + " not seen.\n")
	}
}