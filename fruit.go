package main

import "mud"

type Fruit struct {
	mud.Persister
	mud.PhysicalObject
	mud.TimeListener
	universe *mud.Universe
	room *mud.Room
	name string
	ping chan int
	stageNo int
}

var stages = []string {
	"under-ripe",
	"ripe",
	"rotten",
	"pit" }

func (f Fruit) Visible() bool { return true }
func (f Fruit) Carryable() bool { return true }
func (f Fruit) TextHandles() []string {
	return []string{f.name}
}
func (f Fruit) Description() string {
	return "A " + f.name;
}

func (f *Fruit) SetRoom(r *mud.Room) {
	mud.Log("SetRoom",r)
	f.room = r 
}
func (f *Fruit) Room() *mud.Room { return f.room }

type FruitTasteStimulus struct {
	mud.Stimulus
	f *Fruit
}

func (f *Fruit) Ping() chan int { return f.ping }
func (f *Fruit) Age() {
}
func (f *Fruit) UpdateTimeLoop() {
	for {
		<-f.ping
	}
}

func MakeFruit(u *mud.Universe, name string) *Fruit {
	f := new(Fruit)
	f.universe = u
	f.name = name
	f.ping = make(chan int)

//	u.Persistents = append(u.Persistents, f)
	u.TimeListeners = append(u.TimeListeners, f)

	go f.UpdateTimeLoop()

	return f
}