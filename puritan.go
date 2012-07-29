package main

import ("mud"
	"strings")

type Puritan struct {
	mud.Talker
	mud.NPC
	room *mud.Room
	stimuli chan mud.Stimulus
	id int
}

func (p Puritan) ID() int { return p.id }
func (p Puritan) Name() string { return "Mary Magdalene" }
// Only respond to Talk stimulus to scorn people for cursing
func (p Puritan) DoesPerceive(s mud.Stimulus) bool {
	_, ok := s.(mud.TalkerSayStimulus)
	return ok
}
func (p Puritan) TextHandles() []string {
	return []string { "mary", "mm" }
}

func ContainsAny(s string, subs ...string) bool {
	for _,sub := range(subs) {
		if(strings.Contains(s, sub)) {
			return true
		}
	}
	return false
}

func (p Puritan) HandleStimulus(s mud.Stimulus) {
	scast, ok := s.(mud.TalkerSayStimulus)
	stim := mud.TalkerSay(p, "Wash your mouth out, " + scast.Source().Name())
	if !ok {
		panic("Puritan should only receive TalkerSayStimulus")
	} else {
		text := scast.Text()
		if(ContainsAny(text,
			"shit","piss","fuck",
			"cunt","cocksucker",
			"motherfucker","tits")) {
			p.room.Broadcast(stim)
		}
	}
}
func (p Puritan) StimuliChannel() chan mud.Stimulus {
	return p.stimuli
}
func (p Puritan) Visible() bool { return true }
func (p Puritan) Description() string { return p.Name() }
func (p Puritan) Carryable() bool { return false }