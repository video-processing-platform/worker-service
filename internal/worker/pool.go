package worker

import (
	"log"
	"sync"
)

type Job struct {
	Body       []byte
	RetryCount int
}

type Pool struct {
	workerCount int
	jobs        chan Job
	wg          sync.WaitGroup
}

func NewPool(workerCount int) *Pool {
	return &Pool{
		workerCount: workerCount,
		jobs:        make(chan Job, 100),
	}
}

func (p *Pool) Start(handler func(Job)) {

	for i := 0; i < p.workerCount; i++ {
		p.wg.Add(1)

		go func(workerID int) {
			defer p.wg.Done()

			for job := range p.jobs {
				log.Printf("Worker %d processing job", workerID)

				handler(job)
			}
		}(i)
	}
}

func (p *Pool) Submit(job Job) {
	p.jobs <- job
}

func (p *Pool) Stop() {
	close(p.jobs)
	p.wg.Wait()
}
