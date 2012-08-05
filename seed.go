package main

import ("mud"
	"strconv")

func MakeStupidRooms(universe *mud.Universe) *mud.Room {
	puritan := MakePuritan()
	theBall := Ball{}
	theClock := MakeClock()
	ff := MakeFlipFlop(universe)
	universe.Persistents = []mud.Persister{ff}
	universe.TimeListeners = []mud.TimeListener{theClock}
	ballSlice := []mud.PhysicalObject{theBall, theClock, puritan, ff}
	empty := []mud.PhysicalObject{}

	room := mud.NewBasicRoom(universe, 0, "You are in a bedroom.", ballSlice)
	room.AddPerceiver(puritan)
	room.AddPerceiver(ff)
	room.AddPersistent(ff)
	room2 := mud.NewBasicRoom(universe, 0, "You are in a bathroom.", empty)
	puritan.room = room

	tree := MakeFruitTree(universe, "orange")
	room2.AddPhysObj(tree)
	room2.AddPersistent(tree)
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