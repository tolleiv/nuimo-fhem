package scenes

import (
	"fmt"

	"log"

	"github.com/spf13/viper"
	"github.com/tolleiv/nuimo"
)

type controller struct {
	states  []*state
	current int
}

func NewController() *controller {
	c := &controller{current: 0}

	viper.SetConfigName("scenes")
	viper.AddConfigPath(".")
	viper.ReadInConfig()

	scenes := viper.GetStringMap("scene")

	for scene, _ := range scenes {
		fmt.Printf("Scene %s\n", scene)
		s := NewState(scene)
		props := scenes[scene].(map[interface{}]interface{})
		for prop, _ := range props {
			name, ok := prop.(string)
			if !ok {
				log.Fatal("Config messed up")
			}
			val, ok := props[prop].(string)
			if !ok {
				log.Fatal("Config messed up")
			}
			fmt.Printf("   setting %s=%s\n", name, val)
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
	log.Printf("Nuimo ready to receive events")
	for {
		event := <-events
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
		default:
			log.Printf("Event: %s %x %d", event.Key, event.Raw, event.Value)
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
