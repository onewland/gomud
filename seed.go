package main

import "mud"

func MakePuritan() *Puritan {
	puritan := new(Puritan)
	puritan.id = 100
	puritan.stimuli = make(chan mud.Stimulus, 5)
	go mud.StimuliLoop(puritan)
	return puritan
}

func MakeClock() *HeartbeatClock {
	clock := new(HeartbeatClock)
	clock.tPing = make(chan int)
	return clock
}

func MakeStupidRoom(universe *mud.Universe) *mud.Room {
	puritan := MakePuritan()
	ff := MakeFlipFlop(universe)
	theBall := Ball{}
	theClock := MakeClock()
	universe.Persistents = []mud.Persister{ff}
	universe.TimeListeners = []mud.TimeListener{theClock}
	ballSlice := []mud.PhysicalObject{theBall, theClock, puritan, ff}
	empty := []mud.PhysicalObject{}

	room := mud.NewBasicRoom(universe, 1, "You are in a bedroom.", ballSlice)
	room.AddPerceiver(puritan)
	room.AddPerceiver(ff)
	room2 := mud.NewBasicRoom(universe, 2, "You are in a bathroom.", empty)
	puritan.room = room

	mud.ConnectEastWest(room, room2)

	go room.FanOutBroadcasts()
	go room2.FanOutBroadcasts()
	go room.ActionQueue()
	go room2.ActionQueue()
	go theClock.UpdateTimeLoop()

	return room
}