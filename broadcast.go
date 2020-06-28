// Package broadcast implements a typed, one to many, broadcast mechanism for channels.
// Broadcasts are non-blocking, values are discarded on any receive channels which
// are not able to proceed.
//
//
package broadcast

import (
	"fmt"
	"reflect"
	"sync"
)

// A Group has a send channel and any number of receive channels. A group has a
// type associated with it. The values sent over the send and receive channels
// are all of this type.
type Group struct {
	chanType    reflect.Type
	sendType    reflect.Type
	receiveType reflect.Type
	in          reflect.Value
	out         map[reflect.Value][]reflect.SelectCase
	m           sync.Mutex
}

// NewGroup creates a Group with the associated type of the type of i.
func NewGroup(i interface{}) *Group {
	t := reflect.TypeOf(i)
	g := Group{
		chanType:    reflect.ChanOf(reflect.BothDir, t),
		sendType:    reflect.ChanOf(reflect.SendDir, t),
		receiveType: reflect.ChanOf(reflect.RecvDir, t),
		in:          reflect.MakeChan(reflect.ChanOf(reflect.BothDir, t), 0),
		out:         make(map[reflect.Value][]reflect.SelectCase)}

	go func() {
		for {
			d, ok := g.in.Recv()
			if !ok {
				g.m.Lock()
				for i := range g.out {
					i.Close()
					delete(g.out, i)
				}
				g.m.Unlock()
				return
			}
			g.m.Lock()
			for _, v := range g.out {
				v[0].Send = d
				reflect.Select(v)
			}
			g.m.Unlock()
		}
	}()

	return &g
}

// SendChannel returns the send channel of the group. It can be called more than
// once and will always return the same channel.
// Closing the send channel will close all the receive channels in the Group, and
// the resources used by the Group will be freed.
// When the associated type of the Group is T, the return value is guaranteed
// to be of type chan<- T, and it is safe to perform a type assertion of:
//	ch := g.SendChannel().(chan<- T)
func (g *Group) SendChannel() interface{} {
	return g.in.Convert(g.sendType).Interface()
}

// ReceiveChannel creates and returns a receive channel, with buffer size n. If
// the buffer is full, any values sent on the send channel will be discarded until
// there is space in the buffer.
// When the associated type of the Group is T, the return value is guaranteed
// to be of type <-chan T, and it is safe to perform a type assertion of:
//	ch := g.ReceiveChannel(1).(<-chan T)
func (g *Group) ReceiveChannel(n int) interface{} {
	g.m.Lock()
	defer g.m.Unlock()

	ch := reflect.MakeChan(g.chanType, n)
	rch := ch.Convert(g.receiveType)
	g.out[rch] = []reflect.SelectCase{{Dir: reflect.SelectSend, Chan: ch}, {Dir: reflect.SelectDefault}}
	return rch.Interface()
}

// CloseReceiveChannel closes the receive channel and stops any further values
// being sent on it. It must be used instead of the Close built-in when closing
// the receive channel, otherwise a 'send on closed channel' panic is likely.
// CloseReceiveChannel will return an error if c is not of type <-chan T, where T
// is the type associated with the group.
func (g *Group) CloseReceiveChannel(c interface{}) error {
	t := reflect.TypeOf(c)
	if t != g.receiveType {
		return fmt.Errorf("broadcast.CloseReceiveChannel parameter of type %s, expected type %s", t, g.receiveType)

	}
	g.m.Lock()
	defer g.m.Unlock()
	ch := reflect.ValueOf(c)
	ch.Close()
	delete(g.out, ch)
	return nil
}
