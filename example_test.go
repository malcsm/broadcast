package broadcast_test

import (
	"fmt"

	"github.com/malcsm/broadcast"
)

func Example_int() {
	type mesg struct {
		n int
		s string
	}

	// Pass int(0) to NewGroup to create int broadcast group
	g := broadcast.NewGroup(int(0))

	// Get the send channel
	s := g.SendChannel().(chan<- int)

	// Create a receive channel
	r := g.ReceiveChannel(1).(<-chan int)

	// Send and receive two values
	s <- 1
	fmt.Println(<-r)
	s <- 2
	fmt.Println(<-r)

	close(s)

	// output:
	// 1
	// 2
}

func Example_struct() {
	type mesg struct {
		n int
		s string
	}

	// Pass mesg{} to NewGroup to create mesg broadcast group
	g := broadcast.NewGroup(mesg{})

	// Get the send channel
	s0 := g.SendChannel().(chan<- mesg)

	// Create two receive channels each buffering 5 values
	r1 := g.ReceiveChannel(5).(<-chan mesg)
	r2 := g.ReceiveChannel(5).(<-chan mesg)

	// Send 5 values
	for i := 0; i < 5; i++ {
		s0 <- mesg{i, fmt.Sprintf("message %d", i)}
	}

	// No more values to send
	close(s0)

	// Receive the values from r1
	for v := range r1 {
		fmt.Println(v)
	}

	// Receive the values from r2
	for v := range r2 {
		fmt.Println(v)
	}

	// output:
	// {0 message 0}
	// {1 message 1}
	// {2 message 2}
	// {3 message 3}
	// {4 message 4}
	// {0 message 0}
	// {1 message 1}
	// {2 message 2}
	// {3 message 3}
	// {4 message 4}
}
