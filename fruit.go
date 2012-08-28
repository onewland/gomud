package main

import ("mud"
	"fmt")

type Fruit struct {
	mud.PhysicalObject
	AgingTimeListener
	universe *mud.Universe
	room *mud.Room
	name string
	ping chan int
	stage LifeStage
	lastChange int
	visible bool
	hasMadePlant bool
}

var fruitStages map[int]LifeStage

func init() {
	fruitStages = make(map[int]LifeStage)
	underripe := LifeStage{StageNo: 0, Name: "underripe", StageChangeDelay: 10000}
	ripe := LifeStage{StageNo: 1, Name: "ripe", StageChangeDelay: 40000}
	rotten := LifeStage{StageNo: 2, Name: "rotten", StageChangeDelay: 10000}
        pit := LifeStage{StageNo: 3, Name: "pit", StageChangeDelay: 10000}
	pit.Post = BecomePlant
	defunct := LifeStage{StageNo: 4, Name: "defunct", StageChangeDelay: -1}
	addLs(underripe, fruitStages)
	addLs(ripe, fruitStages)
	addLs(rotten, fruitStages)
	addLs(pit, fruitStages)
	addLs(defunct, fruitStages)
}

func (f Fruit) Visible() bool { return f.visible }
func (f Fruit) Carryable() bool { return true }
func (f Fruit) TextHandles() []string {
	return []string{f.name}
}
func (f Fruit) Description() string {
	return fmt.Sprintf("A(n) %s %s", f.stage.Name, f.name);
}

func (f *Fruit) SetRoom(r *mud.Room) { f.room = r }
func (f Fruit) Room() *mud.Room { return f.room }

type FruitTasteStimulus struct {
	mud.Stimulus
	f *Fruit
}

func (f Fruit) Ping() chan int { return f.ping }
func (f Fruit) LastChange() int { return f.lastChange }
func (f Fruit) LifeStages() map[int]LifeStage { return fruitStages }
func (f Fruit) Stage() LifeStage { return f.stage }
func (f *Fruit) SetStage(l LifeStage) { f.stage = l }
func (f *Fruit) SetStageChanged(now int) { f.lastChange = now }

func BecomePlant(atl AgingTimeListener) {
	f := atl.(*Fruit)

	if(f.Room() != nil) {
		mud.Log("Age MakePlant clause, room =",f.Room())
		p := MakePlant(f.universe, f.name)
		f.Room().AddChild(p)
		p.SetRoom(f.Room())
		
		p.Room().Actions() <- mud.VanishAction{Target: f}
	}
}

func MakeFruit(u *mud.Universe, name string) *Fruit {
	f := new(Fruit)
	f.universe = u
	f.name = name
	f.ping = make(chan int)
	f.stage = fruitStages[0]
	f.visible = true

	u.Add(f)

	go AgeLoop(f)

	return f
}