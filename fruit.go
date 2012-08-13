package main

import ("mud"
	"fmt")

type Fruit struct {
//	mud.Persister
	mud.PhysicalObject
	mud.TimeListener
	universe *mud.Universe
	room *mud.Room
	name string
	ping chan int
	stage fruitStage
	lastChange int
	visible bool
	hasMadePlant bool
}

type fruitStage struct {
	stageNo int
	name string
	// Time at current stage, before next change
	// -1 indicates last stage
	stageChangeDelay int
}

var fruitStages map[int]fruitStage

func addFs(fs fruitStage, fruitStages map[int]fruitStage) {
	fruitStages[fs.stageNo] = fs
}

func init() {
	fruitStages = make(map[int]fruitStage)
	underripe := fruitStage{stageNo: 0, name: "underripe", stageChangeDelay: 10000}
	ripe := fruitStage{stageNo: 1, name: "ripe", stageChangeDelay: 40000}
	rotten := fruitStage{stageNo: 2, name: "rotten", stageChangeDelay: 10000}
        pit := fruitStage{stageNo: 3, name: "pit", stageChangeDelay: 10000}
	defunct := fruitStage{stageNo: 4, name: "defunct", stageChangeDelay: -1}
	addFs(underripe, fruitStages)
	addFs(ripe, fruitStages)
	addFs(rotten, fruitStages)
	addFs(pit, fruitStages)
	addFs(defunct, fruitStages)
}

func (f Fruit) Visible() bool { return f.visible }
func (f Fruit) Carryable() bool { return true }
func (f Fruit) TextHandles() []string {
	return []string{f.name}
}
func (f Fruit) Description() string {
	return fmt.Sprintf("A(n) %s %s", f.stage.name, f.name);
}

func (f *Fruit) SetRoom(r *mud.Room) { f.room = r }
func (f Fruit) Room() *mud.Room { return f.room }

type FruitTasteStimulus struct {
	mud.Stimulus
	f *Fruit
}

func (f *Fruit) Ping() chan int { return f.ping }
func (f *Fruit) Age(now int) {
	if(f.stage.stageChangeDelay > 0) {
		mud.Log("Age next stage clause, room =",f.Room())
		nextStage := (f.stage.stageNo + 1)
		f.stage = fruitStages[nextStage]
		f.lastChange = now
	} else if !f.hasMadePlant {
		mud.Log("Age MakePlant clause, room =",f.Room())
		p := MakePlant(f.universe, f.name)
		f.visible = false
		f.Room().AddChild(p)
		p.SetRoom(f.Room())
		f.hasMadePlant = true
	}
}

func (f *Fruit) UpdateTimeLoop() {
	for {
		now := <- f.ping
		if now > (f.lastChange + f.stage.stageChangeDelay) {
			f.Age(now)
		}
	}
}

func MakeFruit(u *mud.Universe, name string) *Fruit {
	f := new(Fruit)
	f.universe = u
	f.name = name
	f.ping = make(chan int)
	f.stage = fruitStages[0]
	f.visible = true

//	u.Persistents = append(u.Persistents, f)
	u.TimeListeners = append(u.TimeListeners, f)

	go f.UpdateTimeLoop()

	return f
}