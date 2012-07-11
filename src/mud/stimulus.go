package mud

type Stimulus interface {
	StimType() string
	Description(p Perceiver) string
}

type PlayerSayStimulus struct {
	Stimulus
	player *Player
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

func (s PlayerSayStimulus) StimType() string { return "say" }
func (s PlayerSayStimulus) Description(p Perceiver) string {
	playerReceiver, ok := p.(*Player)
	if ok && s.player.id == playerReceiver.id {
		return "You say \"" + s.text + "\"\n"
	} 
	return s.player.name + " said " + "\"" + s.text + "\".\n"
}

func (s PlayerPickupStimulus) StimType() string { return "take" }
func (s PlayerPickupStimulus) Description(p Perceiver) string {
	playerReceiver, ok := p.(*Player)
	if ok && s.player.id == playerReceiver.id {
		return "You took " + s.obj.Description() + "\n"
	}
	return s.player.name + " took " + s.obj.Description() + ".\n"
}