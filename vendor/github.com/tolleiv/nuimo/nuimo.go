// Package nuimo provides an interaction layer for Senic Nuimo devices. It allows to receive user inputs and can write out led pictographs to the LED display.
package nuimo

import (
	"encoding/binary"
	"fmt"
	"strings"
	"time"

	"github.com/currantlabs/ble"
	"github.com/currantlabs/ble/examples/lib/gatt"
	"github.com/currantlabs/ble/linux/hci"
	"github.com/currantlabs/ble/linux/hci/cmd"
	"github.com/mgutz/logxi/v1"
)

const SERVICE_BATTERY_STATUS = "180F"
const SERVICE_DEVICE_INFO = "180A"
const SERVICE_LED_MATRIX = "F29B1523CB1940F3BE5C7241ECB82FD1"
const SERVICE_USER_INPUT = "F29B1525CB1940F3BE5C7241ECB82FD2"

const CHAR_BATTERY_LEVEL = "2A19"
const CHAR_DEVICE_INFO = "2A29"
const CHAR_LED_MATRIX = "F29B1524CB1940F3BE5C7241ECB82FD1"
const CHAR_INPUT_FLY = "F29B1526CB1940F3BE5C7241ECB82FD2"
const CHAR_INPUT_SWIPE = "F29B1527CB1940F3BE5C7241ECB82FD2"
const CHAR_INPUT_ROTATE = "F29B1528CB1940F3BE5C7241ECB82FD2"
const CHAR_INPUT_CLICK = "F29B1529CB1940F3BE5C7241ECB82FD2"

const DIR_LEFT = 0
const DIR_RIGHT = 1
const DIR_UP = 2
const DIR_BACKWARDS = 2
const DIR_DOWN = 3
const DIR_TOWARDS = 3
const DIR_UPDOWN = 4

const CLICK_DOWN = 1
const CLICK_UP = 0

var logger = log.New("nuimo")

type Nuimo struct {
	client ble.Client
	events chan Event
	led    *ble.Characteristic
	bttry  *ble.Characteristic
}

type Event struct {
	Key   string
	Value int64
	Raw   []byte
}

// Connect tried to find nearby devices and connects to them. It tries to reconnect when a timeout interval is passed as first argument.
func Connect(params ...int) (*Nuimo, error) {

	ch := make(chan Event, 100)
	n := &Nuimo{events: ch}
	err := n.reconnect()

	if err != nil {
		logger.Fatal("%s", err)
	}

	if len(params) == 1 && params[0] > 0 {
		go n.keepConnected(params[0])
	}

	return n, err
}

func discoverDevice() (ble.Client, error) {
	logger.Info("Discover")
	filter := func(a ble.Advertisement) bool {
		return strings.ToUpper(a.LocalName()) == "NUIMO"
	}

	// Set connection parameters. Only supported on Linux platform.
	d := gatt.DefaultDevice()
	if h, ok := d.(*hci.HCI); ok {
		if err := h.Option(hci.OptConnParams(
			cmd.LECreateConnection{
				LEScanInterval:        0x0004,    // 0x0004 - 0x4000; N * 0.625 msec
				LEScanWindow:          0x0004,    // 0x0004 - 0x4000; N * 0.625 msec
				InitiatorFilterPolicy: 0x00,      // White list is not used
				PeerAddressType:       0x00,      // Public Device Address
				PeerAddress:           [6]byte{}, //
				OwnAddressType:        0x00,      // Public Device Address
				ConnIntervalMin:       0x0006,    // 0x0006 - 0x0C80; N * 1.25 msec
				ConnIntervalMax:       0x0006,    // 0x0006 - 0x0C80; N * 1.25 msec
				ConnLatency:           0x0000,    // 0x0000 - 0x01F3; N * 1.25 msec
				SupervisionTimeout:    0x0048,    // 0x000A - 0x0C80; N * 10 msec
				MinimumCELength:       0x0000,    // 0x0000 - 0xFFFF; N * 0.625 msec
				MaximumCELength:       0x0000,    // 0x0000 - 0xFFFF; N * 0.625 msec
			})); err != nil {
			logger.Fatal("can't set advertising param: %s", err)
		}
	}
	return gatt.Discover(gatt.FilterFunc(filter))
}

func (n *Nuimo) reconnect() error {
	logger.Info("Reconnect")
	if n.client != nil {
		n.client.ClearSubscriptions()
		n.client.CancelConnection()
	}
	client, err := discoverDevice()
	if err != nil {
		return err
	}
	n.client = client
	return n.discoverServices()
}

func (n *Nuimo) keepConnected(refresh int) {

	for {
		c := make(chan []byte, 1)
		go func() {
			logger.Info("Reading batterie")
			data, err := n.client.ReadCharacteristic(n.bttry)
			if err != nil {
				logger.Error("Error", err)
				// this will cause a reconnect
				return
			}
			c <- data

		}()
		select {
		case data := <-c:
			n.battery(data)
		case <-time.After(30 * time.Second):
			n.send(Event{Key: "disconnected"})
			n.reconnect()
		}
		close(c)
		time.Sleep(time.Duration(refresh) * time.Second)
	}

}

// Events provides access to the events channel which contains the user interaction and battery level events
func (n *Nuimo) Events() <-chan Event {
	return n.events
}

// Display sends the passed byte atrix into the LED display of the Nuimo
func (n *Nuimo) Display(matrix []byte, brightness uint8, timeout uint8) {

	displayMatrix := []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}

	for c, dots := range matrix {
		if c > 10 {
			break
		}
		displayMatrix[c] = dots
	}

	displayMatrix[11] = brightness
	displayMatrix[12] = timeout

	n.client.WriteCharacteristic(n.led, displayMatrix, true)
}

// DisplayMatrix transforms a matrix consisting of 0s and 1s into a byte matrix
func DisplayMatrix(dots ...byte) []byte {
	bytes := []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	var b uint8
	var i uint8
	dotCount := uint8(len(dots))

	for b = 0; b < 11; b++ {
		for i = 0; i < 8; i++ {
			dot := (b * 8) + i
			if dot < dotCount && dots[dot] > 0 {
				bytes[b] |= byte(1) << i
			}
		}

	}

	return bytes
}

//  TODO: make sure we only subscribe to the services we need
func (n *Nuimo) discoverServices() error {
	p, err := n.client.DiscoverProfile(true)
	if err != nil {
		return fmt.Errorf("can't discover services: %s\n", err)
	}

	for _, s := range p.Services {

		switch {
		case s.UUID.Equal(ble.MustParse(SERVICE_DEVICE_INFO)):
			for _, c := range s.Characteristics {
				switch {
				case c.UUID.Equal(ble.MustParse(CHAR_DEVICE_INFO)):
					logger.Info("Info subscribed")
					n.client.Subscribe(c, false, n.info)
				default:
					logger.Warn("Unknown device char", "uuid", c.UUID.String())
					n.client.Subscribe(c, false, n.unknown)
				}
			}
		case s.UUID.Equal(ble.MustParse(SERVICE_BATTERY_STATUS)):
			for _, c := range s.Characteristics {
				switch {
				case c.UUID.Equal(ble.MustParse(CHAR_BATTERY_LEVEL)):
					logger.Info("Battery subscribed")
					n.bttry = c
					n.client.Subscribe(c, false, n.battery)

				default:
					logger.Warn("Unknown battery char", "uuid", c.UUID.String())
					n.client.Subscribe(c, false, n.unknown)
				}
			}
		case s.UUID.Equal(ble.MustParse(SERVICE_USER_INPUT)):
			for _, c := range s.Characteristics {
				switch {
				case c.UUID.Equal(ble.MustParse(CHAR_INPUT_CLICK)):
					n.client.Subscribe(c, false, n.click)
				case c.UUID.Equal(ble.MustParse(CHAR_INPUT_ROTATE)):
					n.client.Subscribe(c, false, n.rotate)
				case c.UUID.Equal(ble.MustParse(CHAR_INPUT_SWIPE)):
					n.client.Subscribe(c, false, n.swipe)
				case c.UUID.Equal(ble.MustParse(CHAR_INPUT_FLY)):
					n.client.Subscribe(c, false, n.fly)
				default:
					logger.Warn("Unknown input characteristik", "uuid", c.UUID.String())
					n.client.Subscribe(c, false, n.unknown)
				}
			}
		case s.UUID.Equal(ble.MustParse(SERVICE_LED_MATRIX)):
			for _, c := range s.Characteristics {
				logger.Info("Led found")
				n.led = c
			}
		default:
			logger.Warn("Unknown service %s", "uuid", s.UUID.String())
		}
	}
	n.send(Event{Key: "connected"})
	return nil
}

// Disconnect closes the connection and drops all subscriptions
func (n *Nuimo) Disconnect() error {
	logger.Warn("Nuimo connection closed")
	//close(n.events)
	return n.client.CancelConnection()
}

func (n *Nuimo) battery(req []byte) {
	uval, _ := binary.Uvarint(req)
	level := int64(uval)
	n.send(Event{Key: "battery", Raw: req, Value: level})
}
func (n *Nuimo) info(req []byte) {
	logger.Info("Info: " + string(req))
}

func (n *Nuimo) click(req []byte) {
	uval, _ := binary.Uvarint(req)
	dir := int64(uval)
	switch dir {
	case CLICK_DOWN:
		n.send(Event{Key: "press", Raw: req})
	case CLICK_UP:
		n.send(Event{Key: "release", Raw: req})
	}
}

func (n *Nuimo) rotate(req []byte) {
	uval := binary.LittleEndian.Uint16(req)
	val := int64(int16(uval))
	n.send(Event{Key: "rotate", Raw: req, Value: val})
}
func (n *Nuimo) swipe(req []byte) {
	uval, _ := binary.Uvarint(req)
	dir := int64(uval)
	n.send(Event{Key: "swipe", Raw: req, Value: dir})

	switch dir {
	case DIR_LEFT:
		n.send(Event{Key: "swipe_left", Raw: req})
	case DIR_RIGHT:
		n.send(Event{Key: "swipe_right", Raw: req})
	case DIR_UP:
		n.send(Event{Key: "swipe_up", Raw: req})
	case DIR_DOWN:
		n.send(Event{Key: "swipe_down", Raw: req})
	}
}
func (n *Nuimo) fly(req []byte) {
	uval, _ := binary.Uvarint(req[0:1])
	dir := int(uval)
	uval, _ = binary.Uvarint(req[2:])
	distance := int64(uval)

	switch dir {
	case DIR_LEFT:
		n.send(Event{Key: "fly_left", Raw: req, Value: distance})
	case DIR_RIGHT:
		n.send(Event{Key: "fly_right", Raw: req, Value: distance})
	case DIR_BACKWARDS:
		n.send(Event{Key: "fly_backwards", Raw: req, Value: distance})
	case DIR_TOWARDS:
		n.send(Event{Key: "fly_towards", Raw: req, Value: distance})
	case DIR_UPDOWN:
		n.send(Event{Key: "fly_updown", Raw: req, Value: distance})
	}
}
func (n *Nuimo) unknown(req []byte) {
	n.send(Event{Key: "unknown", Raw: req})
}

// make sure missing event sinks don't block the client
func (n *Nuimo) send(e Event) {
	go func() { n.events <- e }()
}
