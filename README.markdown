# Gomud - Go MUD!

Gomud is a minimal server meant to handle MU* style games. By default
it supports rooms, players, simple text commands, a global heartbeat,
physical entities, and persists to a Redis back-end.

## Build
Add gomud/ to your `GOPATH` environment variable.

Run `make`.

Alternatively, if you do not have make installed:
Run `go install mud`.

Run `go install mud/simple`.

Run `go build` in the base directory of gomud.

## Usage
On the first run, use `./gomud -seed`. This command flushes the
database and then sets up defaults in "seed.go". From then on, a plain
run of `./gomud` or `./gomud -load` will load the current state of
entities from the database.

The default Redis DB number is 3, but can be specified with `-dbno`.

`gomud` is really a simple prototype for the `mud` package, which contains
the "guts" of the application. For building a new mud, you may want to 
completely rewrite the contents of `mud.go`.

## Concepts
Right now, the implementations of these concepts may not be philosophically
correct as elements settle into place, but they should be mostly accurate.

### Universe
A Universe is a collection of Rooms, PhysicalObjects, Players, and 
any entities that would require interaction with each other. It is the dividing
line between the game world and the technical housekeeping, in matters of 
network and database connectivity.

### PhysicalObject(s)
A PhysicalObject is an object that occupies space and exists at a particular
geographic location. It can be visible or not, carryable or not. Importantly,
all `Player`s are `PhysicalObject`s

### Persister 
A Persister is an instance, of nature undefined, that has 
extemporaneous/dynamic value or values saved to the database.

### Perceivers and Stimuli
Perceivers react to the world and actions around them. Players are themselves
`Perceiver`s and things like speech and entry/exit are delegated by how they
receive Stimuli. Stimuli can be custom-designed and generated in `src/mud` or
in `gomud`.

### Room
Rooms contain PhysicalObjects, Persisters, and Perceivers and persist
themselves. They are connected by `RoomConnection`s which define 2-way exits.

### InterObjectAction(s)
`InterObjectAction`s are necessary when a command or action will affect state
in a way that could cause affect the inputs to some other command or action. 
Since things run concurrently in general, there is a special way for players
to "take" items from a room that forces a sequential ordering. This prevents 
copies of the object being made if two players try to "take" at roughly the
same time, for example. Combat actions are not yet implemeented but would be 
an obvious case for the `InterObjectAction` queue.

### Commands
`Command`s can be global or room-specific. Non-global commands are provided by 
`CommandSource` implementors. A `Command` is simply a function that inputs
a player and string arguments. One can add a command to the global commands
by adding it to the `mud.GlobalCommands` hash (the string key is the command
that the user types).

## Extending 
Per-game additions should not go in the `src/mud`. directory. They should
be in the base `gomud/` directory. Some "template" classes to make building
custom NPCs and objects easier are located in `src/mud/simple`.

Changes which affect the structure of the universe should go in src/mud.

### FlipFlop
`flipflop.go` is an example of how to create a
persistent object that responds to the `PlayerSayStimulus` without
modifying any internal (`src/mud/`) code. It responds to a person saying
"bling set [text]" by changing its description to `[text]`. This change
is persisted in the Redis field `flipFlop:[id]:bling`.

### HeartbeatClock
`clock.go` is an example showing how to create an object
that is dependent on the Heartbeat function. `HeartbeatClock` does 
not implement persistence functions, so it will load if seeded but
not remain after a server restart.
