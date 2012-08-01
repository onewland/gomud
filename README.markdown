Gomud is a simple server meant to handle MU* style games. By default
it supports rooms, players, simple text commands, a global heartbeat,
physical entities, and persists to a Redis back-end.

## Build
Add gomud/ to your `GOPATH` environment variable.

Run `go install mud`.
Run `go build` in the base directory of gomud.

## Usage
On the first run, use `./gomud -seed`. This command flushes the database
and then sets up defaults in "seed.go". From then on, a plain run of ./gomud or 
`./gomud -load` will load the current state of entities from the database.

## Concepts
### PhysicalObject(s)
### Persister
### Perceivers and Stimuli
### Room
### InterObjectAction(s)

## Extending
This section requires more explanation. When public, if you have a question, e-mail
onewland@gmail.com

Per-game additions should not go in the src/mud directory. They should be in the
base gomud/ directory.

### FlipFlop
`FlipFlop` is an example class showing how to create a persistent object that responds
to the `PlayerSayStimulus` without modifying any 