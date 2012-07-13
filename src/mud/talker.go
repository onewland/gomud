package mud 

type Talker interface {
	Name() string
	ID() int
}

func TalkerSay(t Talker, s string) TalkerSayStimulus {
	return TalkerSayStimulus{talker: t, text: s}
}