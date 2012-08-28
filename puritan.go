package main

import ("mud"
	"strings")

func ContainsAny(s string, subs ...string) bool {
	for _,sub := range(subs) {
		if(strings.Contains(s, sub)) {
			return true
		}
	}
	return false
}

func puritanHandleSay(s mud.Stimulus, n *SimpleNPC) {
	scast, ok := s.(mud.TalkerSayStimulus)
	stim := mud.TalkerSay(n, "Wash your mouth out, " + scast.Source().Name())
	if !ok {
		panic("Puritan should only receive TalkerSayStimulus")
	} else {
		text := scast.Text()
		if(ContainsAny(text,
			"shit","piss","fuck",
			"cunt","cocksucker",
			"motherfucker","tits")) {
			n.Room().Broadcast(stim)
		}
	}
}

func MakePuritan(universe *mud.Universe) *SimpleNPC {
	puritan := MakeSimpleNPC(universe)
	puritan.AddStimHandler("say", puritanHandleSay)
	puritan.visible = true
	puritan.name = "Penelope Proper"
	puritan.description = puritan.name
	puritan.carryable = false
	go mud.StimuliLoop(puritan)
	return puritan
}

func (p *SimpleNPC) Commands() map[string]mud.Command {
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
	if p.buyer.Money() >= p.price {
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