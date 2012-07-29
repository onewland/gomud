package main

import ("mud"
	"strconv")

type HeartbeatClock struct { 
	mud.PhysicalObject
	mud.TimeListener
	counter int
	tPing chan int
}

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