package health

type HealthMonitor interface {
	Subscribe() <-chan bool
}

type SimpleHealthMonitor struct {
	subscribers []chan bool
}

var _ HealthMonitor = (*SimpleHealthMonitor)(nil)

// Subscribe implements HealthMonitor.
func (m *SimpleHealthMonitor) Subscribe() <-chan bool {
	ch := make(chan bool)
	m.subscribers = append(m.subscribers, ch)
	return ch
}
