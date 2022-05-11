# forkjoin
Go routines library to implement fan-in and fan-out pattern in Go Lang

## FIO Design

1. Multiplexer spawns N goroutines for N worker added through addWorker method on the Multiplexer 
2. Multiplexer's model is request/response, it return only one response form the worker
3. Each worker needs to 
    * Implement Worker interface and return on result channel. Heartbeat is managed for the worker
    * Exit its work on a signal from Manager on the done channel  
	* Worker only need to implement the actual work
4. The worker (goroutine) is considered unhealthy if the heartbeat is delayed by more than two seconds and is restarted 

## Using FIO

1. [As go library](./usage_go_library.md)
2. [GRPC streaming service](./usage_turnout.md)