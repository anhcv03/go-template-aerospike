package util

import (
	"context"
	"sync"
)

/***************************************************************************************************************/

/* Run all goroutines */
func RunAll(
	tasks []func(...interface{}) error,
	args ...interface{},
) error {
	var wg sync.WaitGroup
	errCh := make(chan error, 1)

	_, cancel := context.WithCancel(context.TODO())
	defer cancel()

	for _, task := range tasks {
		wg.Add(1)

		go func(fn func(...interface{}) error) {
			defer wg.Done()

			err := fn(args...)
			if err != nil {
				select {
				case errCh <- err:
					{
						cancel()
					}

				default:
					{
						// Do nothing
					}
				}
			}
		}(task)
	}

	wg.Wait()

	close(errCh)

	for err := range errCh {
		return err
	}

	return nil
}

/* Run all goroutines with limit */
func RunAllWithLimit(
	tasks []func(...interface{}) error,
	limit int,
	args ...interface{},
) error {
	var wg sync.WaitGroup
	errCh := make(chan error, 1)
	sem := make(chan bool, limit)

	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	for _, task := range tasks {
		wg.Add(1)

		sem <- true

		go func(fn func(...interface{}) error) {
			defer wg.Done()

			defer func() {
				<-sem
			}()

			err := fn(args...)
			if err != nil {
				select {
				case errCh <- err:
					{
						cancel()
					}

				default:
					{
						// Do nothing
					}
				}
			}
		}(task)
	}

	wg.Wait()

	close(errCh)

	for err := range errCh {
		return err
	}

	return nil
}

/***************************************************************************************************************/
