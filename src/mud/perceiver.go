package mud

type PerceiveMap map[string]PhysicalObject
type Perceiver interface {
	ID() int
	DoesPerceive(s Stimulus) bool
	PerceiveList(context PerceiveContext) PerceiveMap
	StimuliChannel() chan Stimulus
	HandleStimulus(s Stimulus)
}


func isPerceiver(o interface{}, ifTrue func(Perceiver)) {
	oAsPerceiver, isPerceiver := o.(Perceiver)
	
	if(isPerceiver) { ifTrue(oAsPerceiver) }
}

func init() {
	containerHelper := new(FlexObjHandlerPair)
	containerHelper.Add = func(fc *FlexContainer, o interface{}) {
		isPerceiver(o, func(Perceiver) {
			fc.AddObjToCategory("Perceivers",o)
		})
	}
	containerHelper.Remove = func(fc *FlexContainer, o interface{}) {
		isPerceiver(o, func(Perceiver) {
			fc.RemoveObjFromCategory("Perceivers",o)
		})
	}
	FlexObjHandlers["Perceivers"] = *containerHelper
}

type PerceiveContext int
const (
	TakeContext PerceiveContext = iota
	LookContext
	InvContext
)

func StimuliLoop(p Perceiver) {
	Log("Starting StimuliLoop",p)
	for {
		nextStimulus := <- p.StimuliChannel()
		if p.DoesPerceive(nextStimulus) {
			p.HandleStimulus(nextStimulus)
		}
	}
}