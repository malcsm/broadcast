# Broadcast

[go.dev reference](https://pkg.go.dev/github.com/malcsm/broadcast?tab=doc)

Package broadcast implements a typed, one to many, broadcast mechanism for channels. Broadcasts are non-blocking, values are discarded on any receive channels which are not able to proceed.

## Example

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