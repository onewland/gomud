package mud

type Command func(p *Player, args[] string)

var GlobalCommands map[string]Command

func init() {
	GlobalCommands = make(map[string]Command)
}