Gomud is a simple server meant to handle MU* style games. By default
it supports rooms, players, simple text commands, a global heartbeat,
physical entities, and persists to a Redis back-end.

On the first run, use ./gomud -seed. This command flushes the database
and then sets up defaults. From then on, a plain run of ./gomud or 
./gomud -load will load the current state of entities from the database.