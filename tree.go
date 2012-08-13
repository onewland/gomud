package main

import ("fmt"
	"mud"
	"strconv")

func init() {
	mud.Loaders["fruitTree"] = LoadFruitTree
	mud.PersistentKeys["fruitTree"] = []string { "id", "fruitName", "nextFlowering" }
	mud.PlayerPerceptions["flower"] = DoesPerceiveFlower
}

type FruitTree struct {
	mud.PhysicalObject
	mud.Persister
	mud.TimeListener
	universe *mud.Universe
	room *mud.Room
	fruitName string
	nextFlowering int
	id int
	ping chan int
}

type TreeFlowerStimulus struct {
	mud.Stimulus
	ft *FruitTree
}

func (s TreeFlowerStimulus) StimType() string { return "flower" }
func (s TreeFlowerStimulus) Description(p mud.Perceiver) string {
	return "The " + s.ft.fruitName + " tree has blossomed.\n"
}

func DoesPerceiveFlower(p mud.Player, s mud.Stimulus) bool { return true }

func (f FruitTree) Visible() bool { return true }
func (f FruitTree) Carryable() bool { return false }
func (f FruitTree) TextHandles() []string {
	return []string{}
}
func (f FruitTree) Description() string {
	return "A " + f.fruitName + " tree";
}
func (f *FruitTree) SetRoom(r *mud.Room) { f.room = r }
func (f *FruitTree) Room() *mud.Room { return f.room }

func (f *FruitTree) Ping() chan int { return f.ping }
func (f *FruitTree) Bloom() {
	mud.Log("Bloom in room",f.room)
	newFruit := MakeFruit(f.universe, f.fruitName)
	f.room.Broadcast(TreeFlowerStimulus{ft: f})
	f.room.AddChild(newFruit)
}

func (f *FruitTree) UpdateTimeLoop() {
	base := 300000
	margin := 1250

	for {
		now := <- f.ping
		if f.nextFlowering == -1 {
			f.nextFlowering = now + randRange(base,margin)
		}
		if now == f.nextFlowering {
			f.nextFlowering = now + randRange(base,margin)
			f.Bloom()
		}
	}
}

func (f *FruitTree) PersistentValues() map[string]interface{} {
	vals := make(map[string]interface{})
	if(f.id > 0) {
		vals["id"] = strconv.Itoa(f.id)
	}
	vals["fruitName"] = f.fruitName
	vals["nextFlowering"] = strconv.Itoa(f.nextFlowering)
	return vals
}

func (f *FruitTree) Save() string {
	outID := f.universe.Store.SaveStructure("fruitTree",f.PersistentValues())
	if(f.id == 0) {
		f.id, _ = strconv.Atoi(outID)
	}
	return outID
}

func (f *FruitTree) DBFullName() string {
	return fmt.Sprintf("fruitTree:%d",f.id)
}

func TreeCount(r *mud.Room) int {
	count := 0
	r.WithPhysObjects(func(p *mud.PhysicalObject) {
		po := *p
		if _, isPlant := po.(*FruitTree); isPlant {
			count += 1
		}
	})
	return count
}


func MakeFruitTree(u *mud.Universe, fruitName string) *FruitTree {
	ft := new(FruitTree)
	ft.universe = u
	ft.fruitName = fruitName
	ft.nextFlowering = -1
	ft.ping = make(chan int)

	u.Persistents = append(u.Persistents, ft)
	u.TimeListeners = append(u.TimeListeners, ft)

	go ft.UpdateTimeLoop()

	return ft
}

func LoadFruitTree(u *mud.Universe, id int) interface{} {
	ft := MakeFruitTree(u, "peach")
	vals := u.Store.LoadStructure(mud.PersistentKeys["fruitTree"],
		mud.FieldJoin(":","fruitTree",strconv.Itoa(id)))
	ft.id = id
	ft.fruitName, _ = vals["fruitName"].(string)
	nextFloweringS, _ := vals["nextFlowering"].(string)
	ft.nextFlowering, _ = strconv.Atoi(nextFloweringS)
	return ft
}