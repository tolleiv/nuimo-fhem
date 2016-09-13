package fhem

import (
	"fmt"
	"net"
	"strings"

	"github.com/Cristofori/kmud/telnet"
)

type Fhem struct {
	Address string
}

func (f *Fhem) Commands(commands <-chan string, output chan<- string) error {
	tn, err := f.connect()
	if err != nil {
		return err
	}
	fmt.Println("Awaiting commands")
	for {
		command, more := <-commands
		if len(command) == 0 {
			continue
		}

		fmt.Printf("Trigger command:%s\n", command)
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
		return nil, err
	}
	tn := telnet.NewTelnet(conn)
	return tn, nil
}
