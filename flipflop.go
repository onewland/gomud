package main

import ("mud"
	"strings"
	"fmt"
	"strconv")

type FlipFlop struct {
	mud.NPC
	mud.Persister
	room *mud.Room
	universe *mud.Universe
	stimuli chan mud.Stimulus
	id int
	lastText string
}

func (f FlipFlop) ID() int { return f.id }
func (f FlipFlop) Name() string { return f.lastText }
// Only respond to Talk stimulus to copy them
func (f FlipFlop) DoesPerceive(s mud.Stimulus) bool {
	_, ok := s.(mud.TalkerSayStimulus)
	return ok
}
func (f FlipFlop) TextHandles() []string { return []string{} }
func (f *FlipFlop) HandleStimulus(s mud.Stimulus) {
	scast, ok := s.(mud.TalkerSayStimulus)
	if !ok {
		panic("FF should only receive TalkerSayStimulus")
	} else {
		args := strings.SplitN(scast.Text()," ",3)
		fmt.Println("FF args:",args)
		if(args[0] == "bling") {
			switch(args[1]) {
			case "set":
				f.lastText = args[2]
			}
		}
	}
}
func (f FlipFlop) StimuliChannel() chan mud.Stimulus {
	return f.stimuli
}
func (f FlipFlop) Visible() bool { return true }
func (f FlipFlop) Description() string { return f.Name() }
func (f FlipFlop) Carryable() bool { return false }
func (f FlipFlop) PersistentValues() map[string]interface{} {
	vals := make(map[string]interface{})
	if(f.ID() > 0) {
		vals["id"] = strconv.Itoa(f.ID())
	}
	vals["bling"] = f.lastText
	return vals
}

func (f *FlipFlop) Save() string {
	outID := f.universe.Store.SaveStructure("flipFlop",f.PersistentValues())	
	if(f.id == 0) {
		f.id, _ = strconv.Atoi(outID)
	}
	return outID
}

func (f *FlipFlop) DBFullName() string {
	return fmt.Sprintf("flipFlop:%d", f.id)
}

func MakeFlipFlop(u *mud.Universe) *FlipFlop {
	ff := new(FlipFlop)
	ff.universe = u
	ff.lastText = "Unchanged."
	ff.stimuli = make(chan mud.Stimulus, 5)
	u.Persistents = append(u.Persistents, ff)
	go mud.StimuliLoop(ff)
	return ff
}

func BuildFFInRoom(u *mud.Universe, p *mud.Player, args []string) {
	ff := MakeFlipFlop(u)
	ff.lastText = strings.Join(args, " ")
	room := p.Room()
	room.AddPerceiver(ff)
	room.AddPhysObj(ff)
	room.AddPersistent(ff)
}