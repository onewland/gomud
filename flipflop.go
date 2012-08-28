package main

import ("mud"
	"strings"
	"fmt"
	"strconv")

func init() {
	mud.Loaders["flipFlop"] = LoadFlipFlop
	mud.PersistentKeys["flipFlop"] = []string{ "id", "bling" }
}

type flipFlopPersister struct {
	mud.Persister
	npc *SimpleNPC
	universe *mud.Universe
}

func ffHandleSay(s mud.Stimulus, n *SimpleNPC) {
	scast, ok := s.(mud.TalkerSayStimulus)
	if !ok {
		panic("FF should only receive TalkerSayStimulus")
	} else {
		args := strings.SplitN(scast.Text()," ",3)
		mud.Log("FF args:",args)
		if(args[0] == "bling") {
			switch(args[1]) {
			case "set":
				n.Meta["lastText"] = args[2]
				n.description = n.Meta["lastText"].(string)
			}
		}
	}
}

func (f flipFlopPersister) PersistentValues() map[string]interface{} {
	vals := make(map[string]interface{})
	if(f.npc.ID() > 0) {
		vals["id"] = strconv.Itoa(f.npc.ID())
	}
	vals["bling"] = f.npc.Meta["lastText"]
	return vals
}
func (f *flipFlopPersister) Save() string {
	outID := f.universe.Store.SaveStructure("flipFlop",f.PersistentValues())
	if(f.npc.id == 0) {
		f.npc.id, _ = strconv.Atoi(outID)
	}
	return outID
}

func (f *flipFlopPersister) DBFullName() string {
	return fmt.Sprintf("flipFlop:%d", f.npc.id)
}

func MakeFlipFlop(u *mud.Universe) *SimpleNPC {
	ff := MakeSimpleNPC(u)
	persister := new(flipFlopPersister)

	persister.npc = ff
	persister.universe = u

	ff.universe = u
	ff.AddStimHandler("say", ffHandleSay)
	ff.Meta["lastText"] = "Unchanged."
	ff.description = ff.Meta["lastText"].(string)
	ff.visible = true

	u.Add(ff)
	u.Add(persister)

	go mud.StimuliLoop(ff)

	return ff
}

func BuildFFInRoom(u *mud.Universe, p *mud.Player, args []string) {
	ff := MakeFlipFlop(u)
	ff.Meta["lastText"] = strings.Join(args, " ")
	room := p.Room()
	room.AddChild(ff)
}

func LoadFlipFlop(u *mud.Universe, id int) interface{} {
	var ok bool
	vals := u.Store.LoadStructure(mud.PersistentKeys["flipFlop"],
		mud.FieldJoin(":","flipFlop",strconv.Itoa(id)))
	ff := MakeFlipFlop(u)
	ff.id = id
	ff.Meta["lastText"], ok = vals["bling"].(string)
	ff.description = ff.Meta["lastText"].(string)
	if !ok { panic("flipFlop:bling not string") }
	return ff
}