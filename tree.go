package main

import ("fmt"
	"mud"
	"math/rand")

func init() {
	mud.Loaders["fruitTree"] = LoadFruitTree
	mud.PersistentKeys["fruitTree"] = []string { "id" }
}

type FruitTree struct {
	mud.PhysicalObject
	mud.Persister
	mud.TimeListener
	universe *mud.Universe
	fruitName string
	fruitDropFreq int
	nextFlowering int
	ping chan int
}

func (f FruitTree) Visible() bool { return true }
func (f FruitTree) Carryable() bool { return false }
func (f FruitTree) TextHandles() []string {
	return []string{}
}
func (f FruitTree) Description() string {
	return "A " + f.fruitName + " tree";
}
func (f *FruitTree) Ping() chan int { return f.ping }
func (f *FruitTree) Bloom() {
	fmt.Println("Flowering")
}

func (f *FruitTree) UpdateTimeLoop() {
	for { 
		now := <- f.ping
		if(now == f.nextFlowering) {
			f.nextFlowering = now + 30000 + (rand.Int()%250)- 
				(rand.Int()%250);
			f.Bloom()
		}
	}
}

func MakeFruitTree(u *mud.Universe, fruitName string) *FruitTree {
	ft := new(FruitTree)
	ft.universe = u
	ft.fruitName = fruitName
	ft.ping = make(chan int)

//	u.Persistents = append(u.Persistents, ft)
	u.TimeListeners = append(u.TimeListeners, ft)

	go ft.UpdateTimeLoop()

	return ft
}

func LoadFruitTree(u *mud.Universe, id int) interface{} {
//	var ok bool
	ft := MakeFruitTree(u, "orange")
	return ft
}