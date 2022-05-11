package forkjoin

import (
	"context"
	"time"
)

//implements the template worker algo for all workers and dispatches work to worker
func dispatch(ctx context.Context, i input, w Worker, pulseInterval time.Duration) (<-chan Result, <-chan heartbeat) {

	pulseStream := make(chan heartbeat)
	resultStream := make(chan Result)
	quitPulstStream := make(chan interface{})

	go func() {
		defer close(quitPulstStream)
		defer close(resultStream)
		result := w.Work(ctx, i.x)
		for {
			select {
			case r := <-result:
				resultStream <- r
				quitPulstStream <- struct{}{}
				return
			default:
			}
		}
	}()

	go func() {
		defer close(pulseStream)
		pulse := time.Tick(pulseInterval)

		sendPulse := func() {
			select {
			case pulseStream <- heartbeat{}:
			default:
			}
		}
		//send pulse at a interval or 1 second
		for {
			select {
			case <-quitPulstStream:
				return
			case <-pulse:
				sendPulse()
			default:
			}
		}
	}()
	return resultStream, pulseStream
}
