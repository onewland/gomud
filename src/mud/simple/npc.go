// simple provides basic implementations for interfaces used in
// package mud.
package simple

import "mud"

type SimpleStimulusHandler func(mud.Stimulus, *NPC)

/*
 Simple, flexible implementation of NPCs
 */
type NPC struct {
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

func (n NPC) ID() int { return n.id }
func (n NPC) SetId(id int) { n.id = id }
func (n NPC) Name() string { return n.name }
func (n NPC) Description() string { return n.description }
func (n *NPC) SetDescription(d string) { n.description = d }
func (n NPC) Carryable() bool { return n.carryable }
func (n *NPC) SetCarryable(c bool) { n.carryable = c}
func (n NPC) Visible() bool { return n.visible }
func (n *NPC) SetVisible(v bool) { n.visible = v }
func (n *NPC) SetUniverse(u *mud.Universe) { n.universe = u }

func (n *NPC) SetRoom(r *mud.Room) { n.room = r }
func (n *NPC) Room() *mud.Room { return n.room }

func (n *NPC) TextHandles() []string { return n.textHandles }

func (n *NPC) AddStimHandler(stimName string, handler SimpleStimulusHandler) {
	mud.Log("AddStimHandler", stimName, handler)
	n.supportedStimuli[stimName] = handler
}

func (n NPC) DoesPerceive(s mud.Stimulus) bool {
	_, there := n.supportedStimuli[s.StimType()]
	return there
}

func (n *NPC) HandleStimulus(s mud.Stimulus) {
	handler := n.supportedStimuli[s.StimType()]
	handler(s, n)
}

func (n NPC) StimuliChannel() chan mud.Stimulus {
	return n.stimuli 
}

func (npc *NPC) AddCommand(text string, c mud.Command) {
	npc.localCommands[text] = c
}

func (npc *NPC) Commands() map[string]mud.Command {
	return npc.localCommands
}

func NewNPC(u *mud.Universe) *NPC {
	npc := new(NPC)
	npc.universe = u
	npc.localCommands = make(map[string]mud.Command)
	npc.supportedStimuli = make(map[string]SimpleStimulusHandler)
	npc.stimuli = make(chan mud.Stimulus, 5)
	npc.Meta = make(map[string]interface{})
	go mud.StimuliLoop(npc)
	return npc
}