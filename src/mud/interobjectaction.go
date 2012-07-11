package mud

type InterObjectAction interface {
	Targets() []PhysicalObject
	Source() PhysicalObject
	Exec()
}