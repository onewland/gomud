package main

import ("mud"
	"strings")

type Puritan struct {
	mud.CommandSource
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

func (p *Puritan) SetRoom(r *mud.Room) { p.room = r }
func (p *Puritan) Room() *mud.Room { return p.room }

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

func MakePuritan() *Puritan {
	puritan := new(Puritan)
	puritan.id = 100
	puritan.stimuli = make(chan mud.Stimulus, 5)
	go mud.StimuliLoop(puritan)
	return puritan
}

func (p *Puritan) Commands() map[string]mud.Command {
	localCommands := make(map[string]mud.Command)
	localCommands["buy"] = buy
	return localCommands
}

type PurchaseAction struct {
	mud.InterObjectAction
	saleObject mud.PhysicalObject
	price mud.Currency
	buyer *mud.Player
}

func (p PurchaseAction) Targets() []mud.PhysicalObject {
	targets := make([]mud.PhysicalObject, 1)
	targets[0] = p.buyer
	return targets
}
func (p PurchaseAction) Source() mud.PhysicalObject { return p.buyer }
func (p PurchaseAction) Exec() {
	if p.buyer.Money() > p.price {
		if p.buyer.ReceiveObject(&p.saleObject) {
			p.buyer.AdjustMoney(-p.price)
			p.buyer.WriteString("Thanks for your purchase!\n\r")
		} else {
			p.buyer.WriteString("You do not have enough space.\n\r")
		}
	} else {
		p.buyer.WriteString("You do not have enough money.\n\r")
	}
}

func buy(p *mud.Player, args[] string) {
	fruit := MakeFruit(p.Universe, args[0])
	action := PurchaseAction{ price: 10, buyer: p, saleObject: fruit }
	p.Room().Actions() <- action
}