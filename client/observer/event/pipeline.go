package event

import "time"

func Process(o Observer, h EventHandler, numberOfEvents int, interval time.Duration) {

	for i := 0; i < numberOfEvents; i++ {
		event := o.GetEvent()
		h.HandleEvent(event)

		time.Sleep(interval)
	}
}
