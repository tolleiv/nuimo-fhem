package scenes

import (
	"fmt"

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
			c.dispatchCommand(c.prevState(), event)
		case "swipe_right":
			c.dispatchCommand(c.nextState(), event)
		case "rotate":
			if event.Value > 10 {
				c.dispatchCommand(c.CurrentState().Handle("rotate_right"), event)
			} else if event.Value < -10 {
				c.dispatchCommand(c.CurrentState().Handle("rotate_left"), event)
			}
		case "press", "release", "swipe_up", "swipe_down":
			c.dispatchCommand(c.CurrentState().Handle(event.Key), event)
		case "swipe":
			// ignore
		case "battery":
			c.dispatchCommand(c.nullState.Handle("battery"), event)
		case "connected", "disconnected":
			c.dispatchCommand(c.nullState.Handle(event.Key), event)
		default:
			logger.Warn(fmt.Sprintf("Unhandled event: %s %x %d", event.Key, event.Raw, event.Value))
			c.dispatchCommand(c.nullState.Handle(event.Key), event)
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

func (c *controller) dispatchCommand(fullCommand string, event nuimo.Event) {

	cmd, _ := NewCommand(fullCommand, event)

	if _, present := c.commandListeners[cmd.handle]; present {
		for _, handler := range c.commandListeners[cmd.handle] {
			go func(handler chan string) {
				handler <- cmd.command
			}(handler)
		}
	}
}
