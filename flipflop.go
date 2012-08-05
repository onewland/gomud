package main

import ("mud"
	"strings"
	"fmt"
	"strconv")

func init() {
	mud.Loaders["flipFlop"] = LoadFlipFlop
	mud.PersistentKeys["flipFlop"] = []string{ "id", "bling" }
}

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
	mud.Log("DoesPerceive entered");
	_, isSay := s.(mud.TalkerSayStimulus)
	return isSay
}
func (f FlipFlop) TextHandles() []string { return []string{} }
func (f *FlipFlop) HandleStimulus(s mud.Stimulus) {
	scast, ok := s.(mud.TalkerSayStimulus)
	if !ok {
		panic("FF should only receive TalkerSayStimulus")
	} else {
		args := strings.SplitN(scast.Text()," ",3)
		mud.Log("FF args:",args)
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
func (f FlipFlop) SetRoom(r *mud.Room) { f.room = r }
func (f FlipFlop) Room() *mud.Room { return f.room }

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

func LoadFlipFlop(u *mud.Universe, id int) interface{} {
	var ok bool
	vals := u.Store.LoadStructure(mud.PersistentKeys["flipFlop"],
		mud.FieldJoin(":","flipFlop",strconv.Itoa(id)))
	ff := MakeFlipFlop(u)
	ff.id = id
	ff.lastText, ok = vals["bling"].(string)
	if !ok { panic("flipFlop:bling not string") }
	return ff
}