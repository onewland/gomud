package mud

type PerceiveMap map[string]PhysicalObject
type Perceiver interface {
	DoesPerceive(s Stimulus) bool
	PerceiveList() PerceiveMap
}