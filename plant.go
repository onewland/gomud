package main

import ("mud"
	"fmt")

type Plant struct {
//	mud.Persister
	mud.PhysicalObject
	mud.TimeListener
	universe *mud.Universe
	room *mud.Room
	name string
	ping chan int
	stage LifeStage
	lastChange int
	hasMadeTree bool
}

var plantStages map[int]LifeStage

func init() {
	plantStages = make(map[int]LifeStage)
	hiddenSeed := LifeStage{StageNo: 0, Name: "hidden-sprout", StageChangeDelay: 10000}
	sprout := LifeStage{StageNo: 1, Name: "sprout", StageChangeDelay: 20000}
	stalk := LifeStage{StageNo: 2, Name: "stalk", StageChangeDelay: 40000}
	miniTree := LifeStage{StageNo: 3, Name: "infant tree", StageChangeDelay: 20000}
	defunct := LifeStage{StageNo: 4, Name: "defunct", StageChangeDelay: -1}
	addLs(hiddenSeed, plantStages)
	addLs(sprout, plantStages)
	addLs(stalk, plantStages)
	addLs(miniTree, plantStages)
	addLs(defunct, plantStages)
}

func (p Plant) Visible() bool { 
	return (p.stage.Name != "hidden-sprout" && 
		p.stage.Name != "defunct")
}
func (p Plant) Carryable() bool { return false }
func (p Plant) TextHandles() []string {
	return []string{p.name}
}
func (p Plant) Description() string {
	return fmt.Sprintf("A %s.\n", p.stage.Name);
}

func (p *Plant) SetRoom(r *mud.Room) { p.room = r }
func (p Plant) Room() *mud.Room { return p.room }

func (p *Plant) Ping() chan int { return p.ping }
func (p *Plant) Age(now int) {
	if(p.stage.StageChangeDelay > 0) {
		nextStage := (p.stage.StageNo + 1)
		p.stage = plantStages[nextStage]
		p.lastChange = now
	} else if !p.hasMadeTree {
		mud.Log("Age MakeTree clause, room =",p.Room())
		if(TreeCount(p.Room()) < 3) {
			t := MakeFruitTree(p.universe, p.name)
			p.Room().AddChild(t)
		}
		p.hasMadeTree = true
	}
}

func (p *Plant) UpdateTimeLoop() {
	for {
		now := <- p.ping
		if now > (p.lastChange + p.stage.StageChangeDelay) {
			p.Age(now)
		}
	}
}

func MakePlant(u *mud.Universe, name string) *Plant {
	p := new(Plant)
	p.universe = u
	p.name = name
	p.ping = make(chan int)
	p.stage = plantStages[0]

//	u.Persistents = append(u.Persistents, f)
	u.TimeListeners = append(u.TimeListeners, p)

	go p.UpdateTimeLoop()

	return p
}