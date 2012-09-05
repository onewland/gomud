package simple

import "mud"

type SimpleTimeHandler func(int, *PhysicalObject)

type PhysicalObject struct {
	mud.PhysicalObject
	mud.TimeListener

	timeHandler SimpleTimeHandler
	tPing chan int
	room *mud.Room
	visible bool
	carryable bool
	description string
	textHandles []string
	universe *mud.Universe
}

func (p *PhysicalObject) Ping() chan int { return p.tPing }
func (p *PhysicalObject) UpdateTimeLoop() {
	for { p.timeHandler(<- p.tPing, p) }
}

func (p *PhysicalObject) SetRoom(r *mud.Room) { p.room = r }
func (p PhysicalObject) Room() *mud.Room { return p.room }

func (p PhysicalObject) Visible() bool { return p.visible }
func (p PhysicalObject) Carryable() bool { return p.carryable }
func (p PhysicalObject) TextHandles() []string { return p.textHandles }
func (p PhysicalObject) Description() string { return p.description }
func (p *PhysicalObject) SetDescription(d string) { p.description = d }
func (p *PhysicalObject) SetCarryable(c bool) { p.carryable = c}
func (p *PhysicalObject) SetVisible(v bool) { p.visible = v }
func (p *PhysicalObject) SetUniverse(u *mud.Universe) { p.universe = u }
func (p *PhysicalObject) SetTextHandles(handles... string) {
	p.textHandles = handles
}
func (p *PhysicalObject) SetTimeHandler(handler SimpleTimeHandler) {
	if p.tPing == nil {
		p.tPing = make(chan int)
	}
	p.timeHandler = handler
}

func NewPhysicalObject(u *mud.Universe) *PhysicalObject {
	p := new(PhysicalObject)
	go p.UpdateTimeLoop()
	return p
}