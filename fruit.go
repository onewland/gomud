package main

import "mud"

type Fruit struct {
	mud.Persister
	mud.PhysicalObject
	mud.TimeListener
	room *mud.Room
	name string
}

func (f *Fruit) SetRoom(r *mud.Room) {
	mud.Log("SetRoom",r)
	f.room = r 
}
func (f *Fruit) Room() *mud.Room { return f.room }

