package scenes

type state struct {
	Name     string
	commands map[string]string
}

func NewState(name string) *state {
	cmds := make(map[string]string)
	return &state{Name: name, commands: cmds}
}

func (s *state) set(cfg string, value string) {
	s.commands[cfg] = value
}

func (s *state) Handle(event string) string {
	logger.Debug("State Handle", s.Name, event, s.commands[event])
	return s.commands[event]
}
