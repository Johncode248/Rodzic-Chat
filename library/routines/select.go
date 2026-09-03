package routines

import (
	"sync"
)

func Select(f ...func() error) error {

	var wg sync.WaitGroup
	wg.Add(len(f))

	var value error
	for _, function := range f {
		go func() {
			defer wg.Done()

			if err := function(); err != nil {
				value = err
			}
		}()
	}

	wg.Wait()

	return value
}
