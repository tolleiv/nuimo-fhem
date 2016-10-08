package scenes

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"strings"

	"github.com/tolleiv/nuimo"
)

type command struct {
	handle  string
	command string
	Value   string
}

func NewCommand(compound string, event nuimo.Event) (*command, error) {
	if strings.TrimSpace(compound) == "" {
		return &command{handle: "empty", command: "", Value: ""}, nil
	}

	parts := strings.SplitN(strings.TrimSpace(compound), ":", 2)
	if len(parts) != 2 {
		return nil, errors.New(fmt.Sprintf("Invalid command %s", compound))
	}

	tmpl, err := template.New("command").Parse(parts[1])
	if err != nil {
		return nil, err
	}
	buf := new(bytes.Buffer)
	tmpl.Execute(buf, event)

	return &command{handle: parts[0], command: buf.String(), Value: ""}, nil
}
