package main

import ("mud"
	"mud/simple"
	"strconv")

func updateDescription(time int, p *simple.PhysicalObject) {
	if time % 100 == 0 {
		p.SetDescription("A large clock reading " + strconv.Itoa(time))
	}
}

func NewClock(universe *mud.Universe) *simple.PhysicalObject {
	clock := simple.NewPhysicalObject(universe)
	clock.SetTimeHandler(updateDescription)
	clock.SetVisible(true)
	clock.SetCarryable(false)
	return clock
}