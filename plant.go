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
	stage plantStage
	lastChange int
	hasMadeTree bool
}

type plantStage struct {
	stageNo int
	name string
	// Time at current stage, before next change
	// -1 indicates last stage
	stageChangeDelay int
}

var plantStages map[int]plantStage

func init() {
	plantStages = make(map[int]plantStage)
	hiddenSeed := plantStage{stageNo: 0, name: "hidden-sprout", stageChangeDelay: 10000}
	sprout := plantStage{stageNo: 1, name: "sprout", stageChangeDelay: 20000}
	stalk := plantStage{stageNo: 2, name: "stalk", stageChangeDelay: 40000}
	miniTree := plantStage{stageNo: 3, name: "infant tree", stageChangeDelay: 20000}
	defunct := plantStage{stageNo: 4, name: "defunct", stageChangeDelay: -1}
	addPs(hiddenSeed, plantStages)
	addPs(sprout, plantStages)
	addPs(stalk, plantStages)
	addPs(miniTree, plantStages)
	addPs(defunct, plantStages)
}

func addPs(ps plantStage, plantStages map[int]plantStage) {
	plantStages[ps.stageNo] = ps
}

func (p Plant) Visible() bool { 
	return (p.stage.name != "hidden-sprout" && 
		p.stage.name != "defunct")
}
func (p Plant) Carryable() bool { return false }
func (p Plant) TextHandles() []string {
	return []string{p.name}
}
func (p Plant) Description() string {
	return fmt.Sprintf("A %s.\n", p.stage.name);
}

func (p *Plant) SetRoom(r *mud.Room) { p.room = r }
func (p Plant) Room() *mud.Room { return p.room }

func (p *Plant) Ping() chan int { return p.ping }
func (p *Plant) Age(now int) {
	if(p.stage.stageChangeDelay > 0) {
		nextStage := (p.stage.stageNo + 1)
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
		if now > (p.lastChange + p.stage.stageChangeDelay) {
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