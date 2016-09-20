package scenes

import (
	"fmt"

	"strings"

	"github.com/mgutz/logxi/v1"
	"github.com/spf13/viper"
	"github.com/tolleiv/nuimo"
)

type controller struct {
	states           []*state
	nullState        *state
	current          int
	commandListeners map[string][]chan string
}

var logger = log.New("nuimo-fhem")

func NewController() *controller {
	c := &controller{current: 0}
	c.commandListeners = make(map[string][]chan string)

	viper.SetConfigName("scenes")
	viper.AddConfigPath(".")
	viper.ReadInConfig()

	defaultScene := viper.GetStringMapString("default")
	logger.Debug("Scene Default")
	c.nullState = NewState("null", defaultScene)

	scenes := viper.GetStringMap("scenes")
	for scene, _ := range scenes {
		logger.Debug("Scene", scene)
		c.appendState(NewState(scene, viper.GetStringMapString("scenes."+scene)))
	}

	return c
}

func (c *controller) appendState(s *state) {
	if c.states == nil {
		c.states = make([]*state, 0, 8)
	}
	c.states = append(c.states, s)
}

func (c *controller) Listen(events <-chan nuimo.Event) {
	logger.Info("Nuimo ready to receive events")
	for {
		event := <-events
		logger.Debug(fmt.Sprintf("Event: %s %x %d", event.Key, event.Raw, event.Value))
		switch event.Key {
		case "swipe_left":
			c.dispatchCommand(c.prevState())
		case "swipe_right":
			c.dispatchCommand(c.nextState())
		case "rotate":
			if event.Value > 10 {
				c.dispatchCommand(c.CurrentState().Handle("rotate_right"))
			} else if event.Value < -10 {
				c.dispatchCommand(c.CurrentState().Handle("rotate_left"))
			}
		case "press", "release", "swipe_up", "swipe_down":
			c.dispatchCommand(c.CurrentState().Handle(event.Key))
		case "swipe":
			// ignore
		case "battery":
			if event.Value > 80 {
				c.dispatchCommand(c.nullState.Handle("battery_ok"))
			} else if event.Value > 40 {
				c.dispatchCommand(c.nullState.Handle("battery_medium"))
			} else {
				c.dispatchCommand(c.nullState.Handle("battery_low"))
			}
		case "connected", "disconnected":
			c.dispatchCommand(c.nullState.Handle(event.Key))
		default:
			logger.Warn(fmt.Sprintf("Unhandled event: %s %x %d", event.Key, event.Raw, event.Value))
			c.dispatchCommand(c.nullState.Handle(event.Key))
		}
	}
}

func (c *controller) CurrentState() *state {
	return c.states[c.current]
}

func (c *controller) nextState() string {
	c.current = (c.current + 1) % len(c.states)
	return c.CurrentState().Handle("id")
}
func (c *controller) prevState() string {
	c.current = (c.current + len(c.states) - 1) % len(c.states)
	return c.CurrentState().Handle("id")
}

func (c *controller) AddCommandListener(prefix string, responseChannel chan string) {
	if _, present := c.commandListeners[prefix]; present {
		c.commandListeners[prefix] =
			append(c.commandListeners[prefix], responseChannel)
	} else {
		c.commandListeners[prefix] = []chan string{responseChannel}
	}
}

func (c *controller) RemoveCommandListener(prefix string, listenerChannel chan string) {
	if _, present := c.commandListeners[prefix]; present {
		for idx, _ := range c.commandListeners[prefix] {
			if c.commandListeners[prefix][idx] == listenerChannel {
				c.commandListeners[prefix] = append(c.commandListeners[prefix][:idx],
					c.commandListeners[prefix][idx+1:]...)
				break
			}
		}
	}
}

func (c *controller) dispatchCommand(fullCommand string) {

	if strings.TrimSpace(fullCommand) == "" {
		return
	}

	parts := strings.SplitN(strings.TrimSpace(fullCommand), ":", 2)
	if len(parts) != 2 {
		logger.Error("Invalid command %s", fullCommand)
	}

	prefix := parts[0]
	command := parts[1]
	logger.Info("Command", fullCommand)
	if _, present := c.commandListeners[prefix]; present {
		for _, handler := range c.commandListeners[prefix] {
			go func(handler chan string) {
				handler <- command
			}(handler)
		}
	}
}
