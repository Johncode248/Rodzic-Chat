package routines

import (
	"sync"
)

func Select(f ...func() error) error {

	var wg sync.WaitGroup
	wg.Add(len(f))

	var mu sync.Mutex
	var value error
	for _, function := range f {
		go func() {
			defer wg.Done()

			if err := function(); err != nil {
				mu.Lock()
				value = err
				mu.Unlock()
			}
		}()
	}

	wg.Wait()

	return value
}
