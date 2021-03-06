package main

import (
	"fmt"

	"flag"

	"github.com/mgutz/logxi/v1"
	"github.com/tolleiv/nuimo"
	"github.com/tolleiv/nuimo-fhem/fhem"
	"github.com/tolleiv/nuimo-fhem/scenes"
)

var logger = log.New("nuimo-fhem")

func main() {

	fhemHost := flag.String("host", "localhost", "Hostname for the FHEM server")
	fhemPort := flag.Int("port", 7072, "Telnet port of the FHEM server")
	nuimoTtl := flag.Int("keepalive", 300, "Nuimo keepalive time in seconds")
	flag.Parse()

	device, _ := nuimo.Connect(*nuimoTtl)
	defer device.Disconnect()

	done := make(chan bool)
	fhemCmds := make(chan string)
	nuimoCmds := make(chan string)
	outTerminal := make(chan string)

	c := scenes.NewController()
	c.AddCommandListener("fhem", fhemCmds)
	c.AddCommandListener("nuimo", nuimoCmds)

	f := &fhem.Fhem{Address: fmt.Sprintf("%s:%d", *fhemHost, *fhemPort)}
	go f.Commands(fhemCmds, outTerminal)

	go func(outputs <-chan string) {
		for {
			out, more := <-outputs
			if !more {
				return
			}
			logger.Info("Fhem output", out)
		}
	}(outTerminal)

	go func(icons <-chan string) {
		for {
			icon := <-icons
			device.Display(iconToMatrix(icon), 255, 10)

		}
	}(nuimoCmds)

	go c.Listen(device.Events())

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
	case "plug":
		matrix = nuimo.DisplayMatrix(
			0, 0, 0, 0, 0, 0, 0, 0, 0,
			0, 0, 0, 1, 1, 1, 0, 0, 0,
			0, 0, 1, 0, 0, 0, 1, 0, 0,
			0, 1, 0, 0, 0, 0, 0, 1, 0,
			0, 1, 0, 1, 0, 1, 0, 1, 0,
			0, 1, 0, 0, 0, 0, 0, 1, 0,
			0, 0, 1, 0, 0, 0, 1, 0, 0,
			0, 0, 0, 1, 1, 1, 0, 0, 0,
			0, 0, 0, 0, 0, 0, 0, 0, 0,
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
