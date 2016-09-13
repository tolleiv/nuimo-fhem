package main

import (
	"fmt"

	"strings"

	"github.com/tolleiv/nuimo"
	"github.com/tolleiv/nuimo-fhem/fhem"
	"github.com/tolleiv/nuimo-fhem/scenes"
)

func main() {

	device, _ := nuimo.Connect()
	defer device.Disconnect()

	done := make(chan bool)
	allCmd := make(chan string)
	fhemCmd := make(chan string)
	nuimoCmd := make(chan string)
	outTerminal := make(chan string)

	c := scenes.NewController()
	go c.Listen(device.Events(), allCmd)

	f := &fhem.Fhem{Address: "loungepi.local:7072"}
	go f.Commands(fhemCmd, outTerminal)

	// dispatch incoming commands
	go func(commands <-chan string, f chan<- string, n chan<- string) {
		for {
			cmd := <-commands
			switch {
			case strings.HasPrefix(cmd, "fhem:"):
				f <- strings.TrimPrefix(cmd, "fhem:")
			case strings.HasPrefix(cmd, "nuimo:"):
				n <- strings.TrimPrefix(cmd, "nuimo:")
			}
		}

	}(allCmd, fhemCmd, nuimoCmd)

	go func(outputs <-chan string) {
		for {
			out, more := <-outputs
			if !more {
				return
			}
			fmt.Printf("Output: %s\n", out)
		}
	}(outTerminal)

	go func(icons <-chan string) {
		for {
			icon := <-icons
			device.Display(iconToMatrix(icon), 255, 10)

		}
	}(nuimoCmd)

	<-done
}

func iconToMatrix(icon string) []byte {

	var matrix []byte

	switch icon {
	case "bulp":
		matrix = nuimo.DisplayMatrix(
			0, 0, 0, 1, 1, 1, 0, 0, 0,
			0, 0, 1, 0, 0, 0, 1, 0, 0,
			0, 1, 0, 0, 0, 0, 0, 1, 0,
			0, 1, 0, 0, 0, 0, 0, 1, 0,
			0, 1, 0, 0, 1, 0, 0, 1, 0,
			0, 0, 1, 0, 0, 0, 1, 0, 0,
			0, 0, 0, 1, 1, 1, 0, 0, 0,
			0, 0, 0, 1, 1, 1, 0, 0, 0,
			0, 0, 0, 0, 1, 0, 0, 0, 0,
		)
	case "media":
		matrix = nuimo.DisplayMatrix(
			0, 0, 1, 1, 1, 1, 1, 1, 0,
			0, 1, 1, 0, 0, 0, 0, 1, 1,
			0, 1, 0, 0, 1, 0, 0, 0, 1,
			0, 1, 0, 0, 1, 1, 0, 0, 1,
			0, 1, 0, 0, 1, 1, 1, 0, 1,
			0, 1, 0, 0, 1, 1, 0, 0, 1,
			0, 1, 0, 0, 1, 0, 0, 0, 1,
			0, 1, 0, 0, 0, 0, 0, 1, 1,
			0, 0, 1, 1, 1, 1, 1, 1, 0,
		)
	case "sound":
		matrix = nuimo.DisplayMatrix(
			0, 0, 0, 0, 0, 0, 0, 0, 0,
			0, 0, 0, 0, 1, 0, 0, 0, 0,
			0, 0, 0, 0, 1, 1, 0, 0, 0,
			0, 0, 0, 0, 1, 0, 1, 0, 0,
			0, 0, 0, 0, 1, 0, 1, 0, 0,
			0, 0, 1, 1, 1, 0, 0, 0, 0,
			0, 1, 0, 0, 1, 0, 0, 0, 0,
			0, 1, 0, 0, 1, 0, 0, 0, 0,
			0, 0, 1, 1, 0, 0, 0, 0, 0,
		)
	case "beamer":
		matrix = nuimo.DisplayMatrix(
			0, 0, 0, 0, 0, 0, 0, 0, 0,
			0, 0, 0, 0, 0, 0, 0, 0, 0,
			0, 0, 0, 0, 0, 0, 0, 0, 0,
			0, 0, 0, 0, 0, 0, 0, 0, 0,
			1, 1, 1, 1, 1, 1, 1, 1, 1,
			1, 1, 1, 1, 1, 1, 0, 0, 1,
			1, 1, 1, 1, 1, 1, 0, 0, 1,
			1, 1, 1, 1, 1, 1, 1, 1, 1,
			0, 1, 1, 0, 0, 0, 1, 1, 0,
		)
	default:
		matrix = nuimo.DisplayMatrix(
			0, 0, 0, 0, 0, 0, 0, 0, 0,
			0, 0, 0, 1, 1, 1, 0, 0, 0,
			0, 0, 0, 0, 0, 0, 1, 0, 0,
			0, 0, 0, 0, 0, 0, 1, 0, 0,
			0, 0, 0, 0, 1, 1, 0, 0, 0,
			0, 0, 0, 0, 1, 0, 0, 0, 0,
			0, 0, 0, 0, 0, 0, 0, 0, 0,
			0, 0, 0, 0, 1, 0, 0, 0, 0,
			0, 0, 0, 0, 0, 0, 0, 0, 0,
		)
	}
	return matrix
}
