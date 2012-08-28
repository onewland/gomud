package mud

type Command func(p *Player, args[] string)

type CommandSource interface {
	Commands() map[string]Command
}

var GlobalCommands map[string]Command

func givesCommands(o interface{}, ifTrue func(CommandSource)) {
	oAsCmdSrc, isCmdSrc := o.(CommandSource)

	if(isCmdSrc) { ifTrue(oAsCmdSrc) }
}

func init() {
	GlobalCommands = make(map[string]Command)

	containerHelper := new(FlexObjHandlerPair)
	containerHelper.Add = func(fc *FlexContainer, o interface{}) {
		givesCommands(o, func(CommandSource) {
			fc.AddObjToCategory("CommandSources",o)
		})
	}
	containerHelper.Remove = func(fc *FlexContainer, o interface{}) {
		givesCommands(o, func(CommandSource) {
			fc.RemoveObjFromCategory("CommandSources",o)
		})
	}
	FlexObjHandlers["CommandSources"] = *containerHelper
}