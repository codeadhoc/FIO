package forkjoin

import (
	"context"
	"fmt"
	"time"
)

//goroutine waits for response from worker on work channel.
func manage(ctx context.Context, i input, multiplexdResultStream chan<- Result) {
	ctx, cancel := context.WithTimeout(ctx, time.Duration(i.worker.ActiveDeadLineSeconds())*time.Second)
	defer cancel()
	defer i.wg.Done()

	log := NewLogger()

	sendResult := func(r Result) {
		multiplexdResultStream <- r
	}

	const pulseInterval = 1 * time.Second
	resultStream, pulseStream := dispatch(ctx, i, i.worker, pulseInterval)
	lastPulseT := time.Now()

	//worker moniter loop
	for {
		select {
		case <-pulseStream:
			currPulseT := time.Now()
			diff := int32(currPulseT.Sub(lastPulseT).Seconds())
			lastPulseT = currPulseT
			if diff > 2 {
				log.LogInfo(fmt.Sprint(i.id), "Heartbeat inconsistent spawning new worker goroutine")
				resultStream, pulseStream = dispatch(ctx, i, i.worker, pulseInterval)
			}
		case r, _ := <-resultStream:
			sendResult(r)
			return
		default:
		}
	}
}
