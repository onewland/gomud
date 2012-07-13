package mud

type PerceiveMap map[string]PhysicalObject
type Perceiver interface {
	ID() int
	DoesPerceive(s Stimulus) bool
	PerceiveList() PerceiveMap
	StimuliChannel() chan Stimulus
	HandleStimulus(s Stimulus)
}

func StimuliLoop(p Perceiver) {
	for {
		nextStimulus := <- p.StimuliChannel()
		if p.DoesPerceive(nextStimulus) {
			p.HandleStimulus(nextStimulus)
		}
	}
}