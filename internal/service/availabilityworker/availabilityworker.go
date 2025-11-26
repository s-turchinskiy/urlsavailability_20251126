// Package availabilityworker Получение доступности url, реализовано на паттерне Worker Pool
package availabilityworker

import (
	"net/http"
	"strings"
	"time"
)

type AvailabilityWorker struct {
	numJobs int
	jobs    chan string
	results chan result

	timeout time.Duration
}

type result struct {
	url       string
	available bool
}

func New(urls []string, rateLimit int, timeout time.Duration) *AvailabilityWorker {

	w := &AvailabilityWorker{
		numJobs: len(urls),
		timeout: timeout,
	}

	w.results = make(chan result, w.numJobs)
	w.jobs = w.generator(urls)

	if rateLimit > w.numJobs {
		rateLimit = w.numJobs
	}

	for i := 1; i <= rateLimit; i++ {
		go w.worker()
	}

	return w

}

func (w *AvailabilityWorker) generator(input []string) chan string {
	jobs := make(chan string, len(input))

	go func() {
		defer close(jobs)

		for _, data := range input {
			jobs <- data
		}
	}()

	return jobs
}

func (w *AvailabilityWorker) worker() {

	for url := range w.jobs {

		client := http.Client{
			Timeout: w.timeout,
		}

		var urlWithProtocol = strings.Clone(url)
		if !strings.Contains(urlWithProtocol, "http://") {
			urlWithProtocol = "http://" + urlWithProtocol
		}
		resp, err := client.Get(urlWithProtocol)

		w.results <- result{
			url:       url,
			available: err == nil && resp != nil && resp.StatusCode == http.StatusOK,
		}
	}
}

func (w *AvailabilityWorker) Result() map[string]bool {

	var rslt result
	results := make(map[string]bool, w.numJobs)
	for a := 1; a <= w.numJobs; a++ {
		select {
		case rslt = <-w.results:
			results[rslt.url] = rslt.available
		}
	}

	close(w.results)

	return results
}
