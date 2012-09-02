package main

import ("mud"
	"strconv")

func MakeStupidRooms(universe *mud.Universe) *mud.Room {
	puritan := NewPuritan(universe)
	theBall := new(Ball)
	theClock := NewClock()
	ff := NewFlipFlop(universe)

	universe.Add(theClock)

	room := mud.NewRoom(universe, 0, "You are in a bedroom.")
	room.AddChild(theBall)
	room.AddChild(theClock)
	room.AddChild(puritan)
	room.AddChild(ff)
	puritan.SetRoom(room)

	room2 := mud.NewRoom(universe, 0, "You are in a bathroom.")

	tree := MakeFruitTree(universe, "peach")
	room2.AddChild(tree)
	tree.room = room2

	src := mud.ConnectEastWest(room, room2)
	universe.Add(src)

	return room
}

func LoadStupidRooms(universe *mud.Universe) *mud.Room {
	roomIds := universe.Store.GlobalSetGet("rooms")
	roomConnIds := universe.Store.GlobalSetGet("roomConnects")
	for _, roomId := range(roomIds) {
		if idNo, err := strconv.Atoi(roomId); err == nil {
			// Note that load room also loads any children
			// persisters
			mud.LoadRoom(universe, idNo)
		} else {
			mud.Log("[warn] strange roomId",roomId)
		}
	}

	for _, roomIdConn := range(roomConnIds) {
		if idNo, err := strconv.Atoi(roomIdConn); err == nil {
			mud.LoadRoomConn(universe, idNo)
		} else {
			mud.Log("[warn] strange roomConnId",roomIdConn)
		}
	}
	return universe.Rooms[1]
}