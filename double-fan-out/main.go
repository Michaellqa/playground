package main

import (
	"sync"
	"time"
)

//

func main() {

}

type T struct {
	Id       int
	Provider string
}

func load(count int) []T {
	res := make([]T, 0, count)
	for i := 0; i < count*4/10; i++ {
		res = append(res, T{Id: i, Provider: "A"})
	}
	for i := 0; i < count*3/10; i++ {
		res = append(res, T{Id: i, Provider: "B"})
	}
	for i := 0; i < count*2/10; i++ {
		res = append(res, T{Id: i, Provider: "C"})
	}
	for i := 0; i < count*1/10; i++ {
		res = append(res, T{Id: i, Provider: "D"})
	}
	return res
}

type Provider struct {
	Name           string
	MaxConnections int
}

type Status struct {
	Id    int
	Value string
}

func (p *Provider) fetch(t T) Status {
	time.Sleep(100 * time.Millisecond)
	return Status{Id: t.Id, Value: p.Name + " ok"}
}

func collectResults(ts []T) []Status {
	// split by providers
	tm := make(map[string][]T)
	for _, t := range ts {
		tm[t.Provider] = append(tm[t.Provider], t)
	}

	wg := sync.WaitGroup{}
	wg.Add(len(tm))

	// fetch for each provider in separate goroutine
	resCh := make(chan []Status, len(tm))
	for k, v := range tm {
		k, v := k, v
		go func() {
			provider := Provider{Name: k}
			resCh <- discoverProvider(provider, v)
			wg.Done()
		}()
	}

	done := make(chan struct{})
	results := make([]Status, 0, len(ts))
	go func() {
		for s := range resCh {
			results = append(results, s...)
		}
		done <- struct{}{}
	}()

	// wait till the last finished
	wg.Wait()
	close(resCh)
	<-done

	return results
}

func discoverProvider(p Provider, ts []T) []Status {
	// producer chan
	srcCh := make(chan T, len(ts))
	for _, t := range ts {
		srcCh <- t
	}

	wg := sync.WaitGroup{}
	wg.Add(len(ts))

	// fan out
	resCh := make(chan Status)
	for i := 0; i < 3; i++ {
		go func() {
			for t := range srcCh {
				resCh <- p.fetch(t)
				wg.Done()
			}
		}()
	}

	// fan in
	done := make(chan struct{})
	res := make([]Status, 0, len(ts))
	go func() {
		for s := range resCh {
			res = append(res, s)
		}
		done <- struct{}{}
	}()

	wg.Wait()
	close(resCh)
	<-done

	return res
}

// try to replace for with buffered chan
