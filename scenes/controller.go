package scenes

import (
	"fmt"

	"github.com/mgutz/logxi/v1"
	"github.com/spf13/viper"
	"github.com/tolleiv/nuimo"
)

type controller struct {
	states    []*state
	nullState *state
	current   int
}

var logger = log.New("nuimo-fhem")

func NewController() *controller {
	c := &controller{current: 0}

	viper.SetConfigName("scenes")
	viper.AddConfigPath(".")
	viper.ReadInConfig()

	defaultScene := viper.GetStringMapString("default")
	c.nullState = NewState("null")
	logger.Info("Scene Default")
	for prop, _ := range defaultScene {
		logger.Info("--->setting", prop, defaultScene[prop])
		c.nullState.set(prop, defaultScene[prop])
	}

	scenes := viper.GetStringMap("scenes")
	for scene, _ := range scenes {
		logger.Info("Scene", scene)
		s := NewState(scene)
		props := scenes[scene].(map[interface{}]interface{})
		for prop, _ := range props {
			name, ok := prop.(string)
			if !ok {
				logger.Fatal("Config messed up")
			}
			val, ok := props[prop].(string)
			if !ok {
				logger.Fatal("Config messed up")
			}
			logger.Info("--->setting", name, val)
			s.set(name, val)
		}

		c.appendState(s)

	}

	return c
}

func (c *controller) appendState(s *state) {
	if c.states == nil {
		c.states = make([]*state, 0, 8)
	}
	c.states = append(c.states, s)
}

func (c *controller) Listen(events <-chan nuimo.Event, commands chan<- string) {
	logger.Info("Nuimo ready to receive events")
	for {
		event := <-events
		logger.Info(fmt.Sprintf("Event: %s %x %d", event.Key, event.Raw, event.Value))
		switch event.Key {
		case "swipe_left":
			commands <- c.prevState()
		case "swipe_right":
			commands <- c.nextState()
		case "rotate":
			if event.Value > 10 {
				commands <- c.CurrentState().Handle("rotate_right")
			} else if event.Value < -10 {
				commands <- c.CurrentState().Handle("rotate_left")
			}
		case "press", "release", "swipe_up", "swipe_down":
			commands <- c.CurrentState().Handle(event.Key)
		case "swipe":
			// ignore
		case "battery":
			if event.Value > 80 {
				commands <- c.nullState.Handle("battery_ok")
			} else if event.Value > 40 {
				commands <- c.nullState.Handle("battery_medium")
			} else {
				commands <- c.nullState.Handle("battery_low")
			}
		case "connected", "disconnected":
			commands <- c.nullState.Handle(event.Key)
		default:
			logger.Warn(fmt.Sprintf("Unhandled event: %s %x %d", event.Key, event.Raw, event.Value))
			commands <- c.nullState.Handle(event.Key)
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
