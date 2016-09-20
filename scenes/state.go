package scenes

type state struct {
	Name     string
	commands map[string]string
}

func NewState(name string, stateCommands map[string]string) *state {
	cmds := make(map[string]string)

	for prop, command := range stateCommands {
		logger.Debug("--->setting", prop, command)
		cmds[prop] = command
	}

	return &state{Name: name, commands: cmds}
}

func (s *state) Handle(event string) string {
	logger.Debug("State Handle", s.Name, event, s.commands[event])
	return s.commands[event]
}
