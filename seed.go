package main

import ("mud"
	"strconv")

func MakeStupidRooms(universe *mud.Universe) *mud.Room {
	puritan := MakePuritan()
	theBall := new(Ball)
	theClock := MakeClock()
	ff := MakeFlipFlop(universe)
	universe.Persistents = []mud.Persister{ff}
	universe.TimeListeners = []mud.TimeListener{theClock}

	room := mud.NewBasicRoom(universe, 0, "You are in a bedroom.")
	room.AddChild(theBall)
	room.AddChild(theClock)
	room.AddChild(puritan)
	room.AddChild(ff)
	puritan.room = room

	room2 := mud.NewBasicRoom(universe, 0, "You are in a bathroom.")

	tree := MakeFruitTree(universe, "peach")
	room2.AddChild(tree)
	tree.room = room2

	src := mud.ConnectEastWest(room, room2)
	universe.AddPersistent(src)

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