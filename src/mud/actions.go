package mud

import "fmt"

type PlayerTakeAction struct {
	InterObjectAction
	player *Player
	target PhysicalObject
	userTargetIdent string
}

func (p PlayerTakeAction) Targets() []PhysicalObject {
	targets := make([]PhysicalObject, 1)
	targets[0] = p.target
	return targets
}
func (p PlayerTakeAction) Source() PhysicalObject { return p.player }
func (p PlayerTakeAction) Exec() {
	fmt.Println("exec take",p.target,p.Source())
	player := p.player
	room := RoomList[player.room]
	if target, ok := player.PerceiveList()[p.userTargetIdent]; ok {
		if target.Carryable() {
			if player.PlaceObjectInInventoryFromRoom(&target, room) {
				room.stimuliBroadcast <- PlayerPickupStimulus{player: player, obj: target}
			} else {
				player.WriteString("No space in your inventory for " + p.userTargetIdent + ".\n")
			}
		} else {
			player.WriteString("Should not take " + p.userTargetIdent + " [not carryable].\n")
		}
	} else {
		player.WriteString(p.userTargetIdent + " not seen.\n")
	}
}