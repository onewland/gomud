package mud

type PhysicalObject interface {
	Visible() bool
	Description() string
	Carryable() bool
	TextHandles() []string
	SetRoom(*Room)
	Room() *Room
}

func ifIsPhysical(o interface{}, ifTrue func(PhysicalObject)) {
	oAsPhysical, isPhysical := o.(PhysicalObject)
	
	if(isPhysical) { ifTrue(oAsPhysical) }
}

func init() {
	containerHelper := new(FlexObjHandlerPair)
	containerHelper.Add = func(fc *FlexContainer, o interface{}) {
		ifIsPhysical(o, func(PhysicalObject) {
			fc.AddObjToCategory("PhysicalObjects",o)
		})
	}
	containerHelper.Remove = func(fc *FlexContainer, o interface{}) {
		ifIsPhysical(o, func(PhysicalObject) {
			fc.RemoveObjFromCategory("PhysicalObjects",o)
		})
	}
	FlexObjHandlers["PhysicalObjects"] = *containerHelper
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