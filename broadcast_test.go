package broadcast_test

import (
	"testing"
	"time"

	"github.com/malcsm/broadcast"
)

func TestBroadcast(t *testing.T) {
	type msg struct {
		i int
	}

	g := broadcast.NewGroup(msg{})
	s0 := g.SendChannel().(chan<- msg)

	// two recipents, each buffering 2 messages
	r1 := g.ReceiveChannel(2).(<-chan msg)
	r2 := g.ReceiveChannel(2).(<-chan msg)

	// broadcast two messages
	s0 <- msg{1}
	s0 <- msg{2}

	// receive the messages
	m := <-r1
	if m.i != 1 {
		t.Fatalf("Recipent 1 received %d expected 1", m.i)
	}

	m = <-r1
	if m.i != 2 {
		t.Fatalf("Recipent 1 received %d expected 2", m.i)
	}

	m = <-r2
	if m.i != 1 {
		t.Fatalf("Recipent 2 received %d expected 1", m.i)
	}

	m = <-r2
	if m.i != 2 {
		t.Fatalf("Recipent 2 received %d expected 2", m.i)
	}

	close(s0)

	m, ok := <-r1
	if ok {
		t.Fatalf("Receiver still open after sender closed")
	}
}

func TestCloseReceive(t *testing.T) {
	type msg struct {
		i int
	}

	g := broadcast.NewGroup(msg{})
	s0 := g.SendChannel().(chan<- msg)

	r1 := g.ReceiveChannel(1).(<-chan msg)

	// broadcast message
	s0 <- msg{1}

	// receive the message
	m := <-r1
	if m.i != 1 {
		t.Fatalf("Recipent received %d expected 1", m.i)
	}

	// close receiver channel
	err := g.CloseReceiveChannel(r1)
	if err != nil {
		t.Fatalf("%s on CloseReceiveChannel", err)
	}

	// broadcast message
	s0 <- msg{2}

	// Check that it wasn't received on the closed channel

	m, ok := <-r1
	if ok {
		t.Errorf("Receiver still open after sender closed")
	}
	if m.i != 0 {
		t.Errorf("Recipent received %d expected 0", m.i)
	}

	// Pass parameter of wrong type
	s := "hello"
	err = g.CloseReceiveChannel(s)
	if err == nil {
		t.Fatalf("No error on wrong parameter type")
	}
	close(s0)
}

func TestNonBlocking(t *testing.T) {
	// Test that sending doesn't block,
	// if a receive channel isn't ready the value should be discarded
	type msg struct {
		i int
	}

	g := broadcast.NewGroup(msg{})
	s0 := g.SendChannel().(chan<- msg)

	// Send a message when no receivers
	w := time.Tick(10 * time.Second)
	select {
	case s0 <- msg{0}:
	case <-w:
		t.Fatalf("Blocking when no receive channel")
	}
	// time.Sleep(time.Second)

	r1 := g.ReceiveChannel(1).(<-chan msg)

	// Send two message to an inactive receiver buffering 1 message
	// 1st message should be received, 2nd discarded
	w = time.Tick(10 * time.Second)
	select {
	case s0 <- msg{1}:
	case <-w:
		t.Fatalf("Blocking on 1st Send")
	}
	w = time.Tick(10 * time.Second)
	select {
	case s0 <- msg{2}:
	case <-w:
		t.Fatalf("Blocking on 2nd Send")
	}
	// time.Sleep(time.Second)

	// The first message should be received
	w = time.Tick(10 * time.Second)
	select {
	case m := <-r1:
		if m.i != 1 {
			t.Fatalf("Recipent received %d expected 1", m.i)
		}
	case <-w:
		t.Fatalf("Blocking on 1st Send")
	}

	// The second message should be discarded
	w = time.Tick(1 * time.Second)
	select {
	case <-r1:
		t.Fatalf("The 2rd message was received, should have been discarded")
	case <-w:
	}

	close(s0)
}

func TestTwosenders(t *testing.T) {

	g := broadcast.NewGroup(0)
	s0 := g.SendChannel().(chan<- int)
	s1 := g.SendChannel().(chan<- int)

	// two recipents, each buffering 2 messages
	r1 := g.ReceiveChannel(2).(<-chan int)
	r2 := g.ReceiveChannel(2).(<-chan int)

	// broadcast two messages
	s0 <- 1
	s1 <- 2

	// receive the messages
	m := <-r1
	if m != 1 {
		t.Fatalf("Recipent 1 received %d expected 1", m)
	}

	m = <-r1
	if m != 2 {
		t.Fatalf("Recipent 1 received %d expected 2", m)
	}

	m = <-r2
	if m != 1 {
		t.Fatalf("Recipent 2 received %d expected 1", m)
	}

	m = <-r2
	if m != 2 {
		t.Fatalf("Recipent 2 received %d expected 2", m)
	}

	close(s0)

	m, ok := <-r1
	if ok {
		t.Fatalf("Receiver still open after sender closed")
	}
}

func TestUnbuffered(t *testing.T) {
	type msg struct {
		i int
	}

	g := broadcast.NewGroup(msg{})
	s0 := g.SendChannel().(chan<- msg)

	r1 := g.ReceiveChannel(0).(<-chan msg)

	go func() {
		s0 <- msg{1}
	}()

	w := time.Tick(10 * time.Second)

	select {
	case m := <-r1:
		if m.i != 1 {
			t.Fatalf("Recipent received %d expected 1", m.i)
		}
	case <-w:
		t.Fatalf("Blocking on Send")
	}
	close(s0)
}
