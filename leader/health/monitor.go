package health

import (
	"fmt"
	"os"
	"time"

	"github.com/base-org/leader-election/leader/config"
)

type HealthMonitor interface {
	Subscribe() <-chan bool
}

type SimpleHealthMonitor struct {
	subscribers   []chan bool
	nodeClient    *Client
	batcherClient *Client
}

var _ HealthMonitor = (*SimpleHealthMonitor)(nil)

func NewSimpleHealthMonitor(cfg *config.Config) HealthMonitor {
	m := &SimpleHealthMonitor{
		subscribers:   make([]chan bool, 0),
		nodeClient:    NewClient(cfg.NodeAddr),
		batcherClient: NewClient(cfg.BatcherAddr),
	}

	go m.notifyHealth()
	return m
}

// Subscribe implements HealthMonitor.
func (m *SimpleHealthMonitor) Subscribe() <-chan bool {
	ch := make(chan bool)
	m.subscribers = append(m.subscribers, ch)
	return ch
}

func (m *SimpleHealthMonitor) notifyHealth() {
	for {
		nodeHealthy, err := m.nodeClient.Healthy()
		if err != nil {
			fmt.Println("node client error", err)
		}

		batcherHealthy, err := m.batcherClient.Healthy()
		if err != nil {
			fmt.Println("batcher client error", err)
		}

		healthy := nodeHealthy && batcherHealthy
		for _, ch := range m.subscribers {
			ch <- healthy
		}

		time.Sleep(2 * time.Second)
	}
}

type MockHealthMonitor struct {
	// Not healthy if file exists, this is used to mock node's health status.
	healthFile  string
	subscribers []chan bool
}

var _ HealthMonitor = (*MockHealthMonitor)(nil)

func NewMockHealthMonitor(healthFile string) HealthMonitor {
	m := &MockHealthMonitor{
		healthFile:  healthFile,
		subscribers: make([]chan bool, 0),
	}

	go m.notifyHealth()
	return m
}

// Subscribe implements HealthMonitor.
func (m *MockHealthMonitor) Subscribe() <-chan bool {
	ch := make(chan bool)
	m.subscribers = append(m.subscribers, ch)
	return ch
}

func (m *MockHealthMonitor) notifyHealth() {
	for {
		healthy := true
		_, err := os.Stat(m.healthFile)
		fmt.Println(m.healthFile, err)
		if err == nil {
			healthy = false
		}

		for _, ch := range m.subscribers {
			ch <- healthy
		}

		time.Sleep(2 * time.Second)
	}
}
