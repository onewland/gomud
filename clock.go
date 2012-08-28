package main

import ("mud"
	"strconv")

type HeartbeatClock struct { 
	mud.PhysicalObject
	mud.TimeListener
	room *mud.Room
	counter int
	tPing chan int
}

func (c *HeartbeatClock) SetRoom(r *mud.Room) { c.room = r }
func (c HeartbeatClock) Room() *mud.Room { return c.room }

func (c HeartbeatClock) Visible() bool { return true }
func (c HeartbeatClock) Carryable() bool { return false }
func (c HeartbeatClock) TextHandles() []string { 
	return []string{"clock", "heartbeat clock"} 
}
func (c HeartbeatClock) Description() string { 
	return "A large clock reading " + strconv.Itoa(c.counter)
}
func (c *HeartbeatClock) Ping() chan int { return c.tPing }
func (c *HeartbeatClock) UpdateTimeLoop() {
	for { c.counter = <- c.tPing }
}

func MakeClock() *HeartbeatClock {
	clock := new(HeartbeatClock)
	clock.tPing = make(chan int)
	go clock.UpdateTimeLoop()
	return clock
}