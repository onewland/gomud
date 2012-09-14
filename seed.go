package main

import ("mud"
	"strconv")

func InitUniverse(universe *mud.Universe) *mud.Room {
	townSquare := mud.NewRoom(universe, 0, 
`Parallax Town Square.

The social and geographic center of the town of Parallax. To the
north is General Seed. The road headed east is Old Town Ave. The
road headed south is Artery Rd.`)
	universe.Add(townSquare)

	oldAve1 := mud.NewRoom(universe, 0, 
`Old Ave.

Old Ave. runs east and west. To the west is town square. There is
a shabby tenement house to the south.`)
	universe.Add(oldAve1)
	mud.ConnectEastWest(townSquare, oldAve1)

	oldAve2 := mud.NewRoom(universe, 0, 
`Old Ave.

Old Ave. runs east and west. To the north is a grocer. To the south
is a church.`)
	universe.Add(oldAve2)
	mud.ConnectEastWest(oldAve1, oldAve2)
	
	oldAve3 := mud.NewRoom(universe, 0,`
Old Ave./Gold St. Intersection

Old Ave. runs west. To the northeast is Gilroy Estate. 
To the north is Gold Street.`)
	universe.Add(oldAve3)
	mud.ConnectEastWest(oldAve2, oldAve3)

	gilroyEstate := mud.NewRoom(universe, 0,`
Gilroy Estate

You are at the entrance to Gilroy Estate. It is a mansion with sprawling
grounds. The garden runs north and east. The foyer is northeast.
`)
	universe.Add(gilroyEstate)
	mud.ConnectNEtoSW(oldAve3, gilroyEstate)

	goldSt1 := mud.NewRoom(universe, 0,`
Gold St.

Gold Street runs south. There is a Patrician Foods to the northeast.`)
	universe.Add(goldSt1)
	mud.ConnectNorthSouth(goldSt1, oldAve3)

	puritan := NewPuritan(universe)
	theBall := NewBall(universe, "&red;red&;")
	theClock := NewClock(universe)
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

	mud.ConnectEastWest(room, room2)

	return townSquare
}

func LoadUniverse(universe *mud.Universe) *mud.Room {
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