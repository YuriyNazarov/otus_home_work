package hw05parallelexecution

import (
	"errors"
	"sync"
)

var ErrErrorsLimitExceeded = errors.New("errors limit exceeded")

type Task func() error

// Run starts tasks in n goroutines and stops its work when receiving m errors from tasks.
func Run(tasks []Task, n, m int) error {
	if n < 1 {
		n = 1
	}
	outChan := make(chan error, n)
	defer close(outChan)
	tasksChan := make(chan Task, n)
	dieFlag := false
	wg := sync.WaitGroup{}
	wg.Add(n)
	mu := sync.Mutex{}

	// create n goroutines
	for i := 0; i < n; i++ {
		go func(dieFlag *bool) {
			for {
				task, ok := <-tasksChan
				mu.Lock()
				die := *dieFlag
				mu.Unlock()
				if !ok || die {
					wg.Done()
					return
				}
				outChan <- task()
			}
		}(&dieFlag)
	}

	// push tasks
	go func(dieFlag *bool) {
		defer close(tasksChan)
		for i := 0; i < len(tasks); i++ {
			mu.Lock()
			die := *dieFlag
			mu.Unlock()
			if die {
				return
			}
			tasksChan <- tasks[i]
		}
	}(&dieFlag)

	// read results
	var errCount int
	for i := 0; i < len(tasks); i++ {
		result := <-outChan
		if result != nil {
			errCount++
		}
		if m > 0 && errCount >= m {
			mu.Lock()
			dieFlag = true
			mu.Unlock()
			wg.Wait()
			return ErrErrorsLimitExceeded
		}
	}
	return nil
}
