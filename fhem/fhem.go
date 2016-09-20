package fhem

import (
	"net"
	"strings"

	"github.com/Cristofori/kmud/telnet"
	"github.com/mgutz/logxi/v1"
)

var logger = log.New("fhem")

type Fhem struct {
	Address string
}

func (f *Fhem) Commands(commands <-chan string, output chan<- string) error {
	tn, err := f.connect()
	if err != nil {
		return err
	}
	logger.Info("Awaiting commands")
	for {
		command, more := <-commands
		if len(command) == 0 {
			continue
		}

		logger.Debug("Trigger command", command)
		if !more {
			break
		}

		data := []byte(command + "\n")

		_, err := tn.Write(data)
		if err != nil {
			return err
		}

		if strings.HasPrefix(command, "get ") || strings.HasPrefix(command, "{ReadingsVal") {
			readBuffer := make([]byte, 1024)
			n, err := tn.Read(readBuffer)
			if err != nil {
				return err
			}
			output <- string(readBuffer[:n])
		} else {
			output <- "ok"
		}
	}
	return nil
}

func (fhem *Fhem) connect() (*telnet.Telnet, error) {
	conn, err := net.Dial("tcp", fhem.Address)
	if err != nil {
		logger.Error("Unable to connect to telnet server", err)
		return nil, err
	}
	tn := telnet.NewTelnet(conn)
	return tn, nil
}
