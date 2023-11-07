package health

import "time"

type HealthMonitor interface {
	Subscribe() <-chan bool
}

type SimpleHealthMonitor struct {
	subscribers []chan bool
}

var _ HealthMonitor = (*SimpleHealthMonitor)(nil)

func NewSimpleHealthMonitor() HealthMonitor {
	m := &SimpleHealthMonitor{
		subscribers: make([]chan bool, 0),
	}
	m.notifyHealth()
	return m
}

// Subscribe implements HealthMonitor.
func (m *SimpleHealthMonitor) Subscribe() <-chan bool {
	ch := make(chan bool)
	m.subscribers = append(m.subscribers, ch)
	return ch
}

func (m *SimpleHealthMonitor) notifyHealth() {
	go func() {
		for {
			for _, ch := range m.subscribers {
				ch <- true
			}
			time.Sleep(2 * time.Second)
		}
	}()
}
