## Usage & Test

1. As go library

Work will be done in 'Work' method which needs to be implemented by worker. Work will be executed concurrently and response result will be returned on a read channel.

Refer [forkjoin_test.go](./forkjoin_test.go) forkjoin_test.go for complete code

```
func TestChecker(t *testing.T) {
	//client can cancel entire processing
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var pc prospectcompany = prospectcompany{}
	m := NewMultiplexer()
	m.AddWorker(&centralbankchecker{})
	resultStream := m.Multiplex(ctx, pc)
	for r := range resultStream {
		if r.Err != nil {
			t.Errorf("error not expected id: %v code: %v message: %v\n", r.ID, r.Err.Code, r.Err.Message)
		} else {
			pc, ok := r.X.(prospectcompany)
			if !ok {
				t.Errorf("type assertion err prospectcompany not found in response")
			} else {
				if pc.isMatch == false {
					t.Errorf("type assertion err prospectcompany not found in response")
				}
			}
		}
	}
}

//worker
func (c *centralbankchecker) Work(done <-chan interface{}, x interface{}) <-chan Result {
	resultStream := make(chan Result)

	go func() {
		defer close(resultStream)
		pc, ok := x.(prospectcompany)
		if !ok {
			resultStream <- Result{Err: &FJerror{Code: RequestError, Message: "type assertion err prospectcompany not found"}}
			return
		}
		n := randInt(15)
		log.Printf("Sleeping %d seconds...\n", n)
		for {
			select {
			case <-done:
				return
			case <-time.After((time.Duration(n) * time.Second)):
				pc.isMatch = true
				resultStream <- Result{X: pc}
				return
			}
		}
	}()
	return resultStream
}
```