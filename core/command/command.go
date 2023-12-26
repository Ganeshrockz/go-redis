package command

func RegisterCommands(r CommandRegistry) {
	RegisterGetCommand(r)
	RegisterSetCommand(r)
	RegisterDeleteCommand(r)
}
