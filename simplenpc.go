package main

import "mud"

type SimpleStimulusHandler func(mud.Stimulus, *SimpleNPC)

type SimpleNPC struct {
	mud.CommandSource
	mud.Talker
	mud.NPC
	name string
	room *mud.Room
	localCommands map[string]mud.Command
	universe *mud.Universe
	stimuli chan mud.Stimulus
	supportedStimuli map[string]SimpleStimulusHandler
	textHandles []string
	id int
	visible bool
	carryable bool
	description string
	Meta map[string]interface{}
}

func (n SimpleNPC) ID() int { return n.id }
func (n SimpleNPC) Name() string { return n.name }
func (n SimpleNPC) Description() string { return n.description }
func (n SimpleNPC) Carryable() bool { return n.carryable }
func (n SimpleNPC) Visible() bool { return n.visible }

func (n *SimpleNPC) SetRoom(r *mud.Room) { n.room = r }
func (n *SimpleNPC) Room() *mud.Room { return n.room }

func (n *SimpleNPC) AddStimHandler(stimName string, handler SimpleStimulusHandler) {
	mud.Log("AddStimHandler", stimName, handler)
	n.supportedStimuli[stimName] = handler
}

func (n SimpleNPC) DoesPerceive(s mud.Stimulus) bool {
	_, there := n.supportedStimuli[s.StimType()]
	return there
}

func (n *SimpleNPC) HandleStimulus(s mud.Stimulus) {
	handler := n.supportedStimuli[s.StimType()]
	handler(s, n)
}

func (n SimpleNPC) StimuliChannel() chan mud.Stimulus {
	return n.stimuli 
}

func (npc *SimpleNPC) Commands() map[string]mud.Command {
	return npc.localCommands
}

func MakeSimpleNPC(u *mud.Universe) *SimpleNPC {
	npc := new(SimpleNPC)
	npc.universe = u
	npc.localCommands = make(map[string]mud.Command)
	npc.supportedStimuli = make(map[string]SimpleStimulusHandler)
	npc.stimuli = make(chan mud.Stimulus, 5)
	npc.Meta = make(map[string]interface{})
	go mud.StimuliLoop(npc)
	return npc
}