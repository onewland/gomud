package main

import ("mud"
	"fmt")

func init() {
	mud.GlobalCommands["pioneer"] = Pioneer
	mud.GlobalCommands["rewrite"] = Rewrite
}

func Pioneer(p *mud.Player, args[] string) {
	direction := args[0]

	fmt.Println("Pioneer",args)
	
	p.Room().WithExit(direction, func(rei *mud.RoomExitInfo) {
		p.WriteString("That exit already exists.\n")
		return
	}, func() {
		BuildPioneerRoom(p, direction)
	})
}

func BuildPioneerRoom(p *mud.Player, direction string) {
	var roomConn *mud.SimpleRoomConnection
	newRoom := mud.NewBasicRoom(p.Universe, 
		0,
		"A default room text.",
		[]mud.PhysicalObject{})
	switch direction {
	case "east":
		roomConn = mud.ConnectEastWest(p.Room(), newRoom)
	case "west":
		roomConn = mud.ConnectEastWest(newRoom, p.Room())
	case "north":
		roomConn = mud.ConnectNorthSouth(p.Room(), newRoom)
	case "south":
		roomConn = mud.ConnectNorthSouth(newRoom, p.Room())
	default:
		p.WriteString("Pioneering only east/west/north/south supported.\n")
		return
	}
	p.Universe.AddPersistent(roomConn)
}

func Rewrite(p *mud.Player, args[] string) {
}