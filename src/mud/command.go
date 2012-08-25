package mud

type Command func(p *Player, args[] string)

type CommandSource interface {
	Commands() map[string]Command
}

var GlobalCommands map[string]Command

func init() {
	GlobalCommands = make(map[string]Command)
}