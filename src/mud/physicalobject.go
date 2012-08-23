package mud

type PhysicalObject interface {
	Visible() bool
	Description() string
	Carryable() bool
	TextHandles() []string
	SetRoom(*Room)
	Room() *Room
}

type VanishAction struct {
	InterObjectAction
	Target PhysicalObject
}

func (p VanishAction) Targets() []PhysicalObject {
	targets := make([]PhysicalObject, 1)
	targets[0] = p.Target
	return targets
}
func (p VanishAction) Source() PhysicalObject { return p.Target }
func (p VanishAction) Exec() {
	plant := p.Target
	room := p.Target.Room()
	room.RemoveChild(plant)
	Log(p, "vanishing")
}