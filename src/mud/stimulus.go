package mud

type Stimulus interface {
	StimType() string
	Description(p Perceiver) string
}

type TalkerSayStimulus struct {
	Stimulus
	talker Talker
	text string
}

type PlayerEnterStimulus struct {
	Stimulus
	player *Player
	from string
}

type PlayerLeaveStimulus struct {
	Stimulus
	player *Player
	to string
}

type PlayerPickupStimulus struct {
	Stimulus
	player *Player
	obj PhysicalObject
}

func (s PlayerEnterStimulus) StimType() string { return "enter" }
func (s PlayerEnterStimulus) Description(p Perceiver) string {
	return s.player.name + " has entered the room.\n"
}

func (s PlayerLeaveStimulus) StimType() string { return "exit" }
func (s PlayerLeaveStimulus) Description(p Perceiver) string {
	return s.player.name + " has left the room.\n"
}

func (s TalkerSayStimulus) StimType() string { return "say" }
func (s TalkerSayStimulus) Description(p Perceiver) string {
	playerReceiver, ok := p.(*Player)
	if ok && s.talker.ID() == playerReceiver.id {
		return "You say \"" + s.text + "\"\n"
	} 
	return s.talker.Name() + " said " + "\"" + s.text + "\".\n"
}
func (s TalkerSayStimulus) Text() string { return s.text }
func (s TalkerSayStimulus) Source() Talker { return s.talker }

func (s PlayerPickupStimulus) StimType() string { return "take" }
func (s PlayerPickupStimulus) Description(p Perceiver) string {
	playerReceiver, ok := p.(*Player)
	if ok && s.player.ID() == playerReceiver.id {
		return "You picked up \"" + s.obj.Description() + "\"\n"
	} 
	return s.player.name + " said " + "\"" + s.obj.Description() + "\".\n"
}