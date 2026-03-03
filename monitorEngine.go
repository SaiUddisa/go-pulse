package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"sync/atomic"
	"time"

	"golang.org/x/sync/errgroup"
)

type Doctor struct {
	Workers int
	Timeout time.Duration
}
type API struct {
	URL        string                 `json:"url"`
	MethodType string                 `json:"method_type"`
	Body       map[string]interface{} `json:"body" omitempty`
}
type Request struct {
	URLs []API `json:"urls"`
}
type Response struct {
	Status  string   `json:"status"`
	Message string   `json:"message"`
	Meta    []string `json:"meta"`
}
type Option func(*Doctor)

func WithTimeout(duration int) Option {
	return func(m *Doctor) {
		m.Timeout = time.Duration(duration) * time.Second
	}
}
func WithWorkers(n int) Option {
	return func(m *Doctor) {
		m.Workers = n
	}
}

func CreateDoctor(opts ...Option) *Doctor {
	m := &Doctor{
		Workers: 2,
		Timeout: 10 * time.Second,
	}
	for _, opt := range opts {
		opt(m)
	}
	return m
}

func (d *Doctor) CheckHealth(ctx context.Context, apis []API) Response {
	//errgroup creation
	g, ctx := errgroup.WithContext(ctx)
	log := getLogger(ctx)
	jobs := make(chan API)
	metaResults := make(chan string)
	var metadata []string
	var processedCount int64
	//Fan-In producer
	g.Go(func() error {
		defer close(jobs)
		for _, api := range apis {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case jobs <- api:
			}
		}
		return nil
	})

	//Fan-out Consumers
	for i := 0; i < d.Workers; i++ {
		g.Go(func() error {
			for api := range jobs {
				select {
				case <-ctx.Done():
					return ctx.Err()
				default:
					// metadata = append(metadata, checkURL(api, d.Timeout, i))
					metaResults <- checkURL(api, d.Timeout, i)
					atomic.AddInt64(&processedCount, 1)
				}
			}
			return nil

		})
	}
	//to collect the meta data
	go func() {

		err := g.Wait()
		if err != nil {
			log.Error("Error Checking Health", slog.String("error", err.Error()))
		}
		close(metaResults)

	}()

	for str := range metaResults {
		metadata = append(metadata, str)
	}

	log.Info("Health Checking Completed!!!\nTotal URLs checked: ", slog.Int64("count", atomic.LoadInt64(&processedCount)))
	return Response{
		Status:  "Success",
		Message: fmt.Sprintf("Health Checking Completed , Total URLs checked : %d", atomic.LoadInt64(&processedCount)),
		Meta:    metadata,
	}

}

func checkURL(api API, timeout time.Duration, workedId int) string {
	client := http.Client{Timeout: timeout}
	switch api.MethodType {
	case "GET":
		res, err := client.Get(api.URL)
		if err != nil {
			return fmt.Sprintf("Error Checking %s: %v\n", api.URL, err)

		}
		defer res.Body.Close()
		return fmt.Sprintf("Worker %d says: [OK] %s: %d", workedId, api.URL, res.StatusCode)
	case "POST":
		jsonData, err := json.Marshal(api.Body)
		if err != nil {
			return fmt.Sprintf("Failed to marshal body for %s: %v\n", api.URL, err)

		}
		body := bytes.NewBuffer(jsonData)
		res, err := client.Post(api.URL, "application/json", body)
		if err != nil {
			fmt.Sprintf("Error Checking %s: %v\n", api.URL, err)
			return ""
		}
		defer res.Body.Close()
		return fmt.Sprintf("Worker %d says: [OK] %s: %d", workedId, api.URL, res.StatusCode)

	default:
		return fmt.Sprintf("Method Not Supported")
	}

}
