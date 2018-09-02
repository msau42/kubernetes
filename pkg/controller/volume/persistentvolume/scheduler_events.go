/*
Copyright 2018 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package persistentvolume

import (
	"math/rand"
	"sync"

	"github.com/golang/glog"
)

// Broadcaster is an interface that allows multiple
// listeners to receive notifications for events
// through a channel
type Broadcaster interface {
	// Register returns a unique id and a channel for
	// listening to events
	Register() (int, chan int)
	// Unregister removes the listener and its channel
	Unregister(int)
	// Broadcast sends a notification to all listeners
	Broadcast()
}

type broadcaster struct {
	listeners map[int]chan int
	mutex     sync.RWMutex
}

func NewBroadcaster() Broadcaster {
	return &broadcaster{
		listeners: map[int]chan int{},
		mutex:     sync.RWMutex{},
	}
}

func (b *broadcaster) Register() (int, chan int) {
	c := make(chan int, 1)

	b.mutex.Lock()
	defer b.mutex.Unlock()

	// Find an unused id
	id := rand.Int()
	for {
		if _, ok := b.listeners[id]; !ok {
			b.listeners[id] = c
			break
		}
		id = rand.Int()
	}

	return id, c
}

func (b *broadcaster) Unregister(id int) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	ch, ok := b.listeners[id]
	if ok {
		close(ch)
		delete(b.listeners, id)
	}
}

func (b *broadcaster) Broadcast() {
	b.mutex.RLock()
	defer b.mutex.RUnlock()

	for id, c := range b.listeners {
		select {
		case c <- 1:
			glog.V(5).Infof("Sent to id %v", id)
		default:
			glog.V(5).Infof("Buffer full for id %v", id)
		}
	}
}
