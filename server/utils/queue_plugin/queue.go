package queue

import (
	"errors"
	"sync"
	"time"
)

type queue struct {
	exit         chan bool
	capacity     int
	topics       map[string][]chan any // key： topic  value ： queue
	sync.RWMutex                       // 同步锁
	//once         sync.Once
}

func NewQueue() *queue {
	return &queue{
		exit:   make(chan bool),
		topics: make(map[string][]chan any),
	}
}

func (q *queue) ShowExit() chan bool {
	return q.exit

}

func (q *queue) SetConditions(capacity int) {
	q.capacity = capacity
}

func (q *queue) Start() {
	select {
	case <-q.exit:
		q.exit = make(chan bool)
	default:
		return
	}
}
func (q *queue) Close() {
	select {
	case <-q.exit:
		return
	default:
		close(q.exit)
		q.Lock()
		q.topics = make(map[string][]chan any)
		q.Unlock()
	}
	return
}

func (q *queue) Publish(topic string, pub any) error {
	select {
	case <-q.exit:
		return errors.New("queue is closed")
	default:
	}
	q.RLock()
	subscribers, ok := q.topics[topic]
	q.RUnlock()
	if !ok {
		return nil
	}
	q.Broadcast(pub, subscribers)
	return nil
}

func (q *queue) Broadcast(msg any, subscribers []chan any) {
	count := len(subscribers)
	concurrency := 1
	switch {
	case count > 1000:
		concurrency = 3
	case count > 100:
		concurrency = 2
	default:
		concurrency = 1
	}
	pub := func(start int) {
		idleDuration := 5 * time.Millisecond
		ticker := time.NewTicker(idleDuration)
		defer ticker.Stop()
		for j := start; j < count; j += concurrency {
			select {
			case subscribers[j] <- msg:
			case <-ticker.C:
			case <-q.exit:
				return
			}
		}
	}
	for i := 0; i < concurrency; i++ {
		go pub(i)
	}
}

func (q *queue) Subscribe(topic string) (<-chan any, error) {
	select {
	case <-q.exit:
		return nil, errors.New("queue is closed")
	default:
	}
	if q.capacity == 0 {
		q.capacity = 100
	}
	ch := make(chan any, q.capacity)
	q.Lock()
	q.topics[topic] = append(q.topics[topic], ch)
	q.Unlock()
	return ch, nil
}

func (q *queue) Unsubscribe(topic string, sub <-chan any) error {
	select {
	case <-q.exit:
		return errors.New("queue is closed")
	default:
	}
	q.RLock()
	subscribers, ok := q.topics[topic]
	q.RUnlock()
	if !ok {
		return nil
	}
	// delete subscriber
	q.Lock()
	var newSubs []chan any
	for _, subscriber := range subscribers {
		if subscriber == sub {
			continue
		}
		newSubs = append(newSubs, subscriber)
	}
	q.topics[topic] = newSubs
	q.Unlock()
	return nil
}

func (q *queue) GetPayLoad(sub <-chan any) any {
	for val := range sub {
		if val != nil {
			return val
		}
	}
	return nil
}
