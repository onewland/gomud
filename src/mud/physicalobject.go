package mud

type PhysicalObject interface {
	Visible() bool
	Description() string
	Carryable() bool
	TextHandles() []string
	SetRoom(*Room)
	Room() *Room
}