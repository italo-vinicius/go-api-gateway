package main

import (
	"context"
	"fmt"
	"github.com/italo/go-api-gateway/internal/app"
	"github.com/italo/go-api-gateway/internal/config"
	"github.com/spf13/cobra"
	"log/slog"
	"net/http"
	"os"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

const version = "0.1.0"

func main() {
	root := &cobra.Command{Use: "go-api-gateway", SilenceUsage: true}
	var path string
	serve := &cobra.Command{Use: "serve", RunE: func(_ *cobra.Command, _ []string) error {
		c, e := config.Load(path)
		if e != nil {
			return e
		}
		return app.Serve(context.Background(), c, slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: parseLevel(c.Logging.Level)})))
	}}
	serve.Flags().StringVar(&path, "config", "configs/go-api-gateway.example.yaml", "YAML configuration")
	var validPath string
	validate := &cobra.Command{Use: "validate", RunE: func(_ *cobra.Command, _ []string) error {
		_, e := config.Load(validPath)
		if e == nil {
			fmt.Println("configuration is valid")
		}
		return e
	}}
	validate.Flags().StringVar(&validPath, "config", "configs/go-api-gateway.example.yaml", "YAML configuration")
	root.AddCommand(serve, validate, &cobra.Command{Use: "version", Run: func(*cobra.Command, []string) { fmt.Println(version) }}, loadCommand())
	if e := root.Execute(); e != nil {
		os.Exit(1)
	}
}
func parseLevel(s string) slog.Level {
	var l slog.Level
	if e := l.UnmarshalText([]byte(s)); e != nil {
		return slog.LevelInfo
	}
	return l
}
func loadCommand() *cobra.Command {
	var target string
	var total, concurrency int
	var timeout time.Duration
	c := &cobra.Command{Use: "load-test", RunE: func(_ *cobra.Command, _ []string) error {
		if total < 1 || concurrency < 1 {
			return fmt.Errorf("requests and concurrency must be positive")
		}
		client := &http.Client{Timeout: timeout}
		start := time.Now()
		var ok, fail atomic.Int64
		lat := make([]time.Duration, 0, total)
		var mu sync.Mutex
		jobs := make(chan struct{})
		var wg sync.WaitGroup
		for range concurrency {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for range jobs {
					s := time.Now()
					r, e := client.Get(target)
					d := time.Since(s)
					mu.Lock()
					lat = append(lat, d)
					mu.Unlock()
					if e == nil && r.StatusCode < 400 {
						ok.Add(1)
					} else {
						fail.Add(1)
					}
					if r != nil {
						r.Body.Close()
					}
				}
			}()
		}
		for range total {
			jobs <- struct{}{}
		}
		close(jobs)
		wg.Wait()
		sort.Slice(lat, func(i, j int) bool { return lat[i] < lat[j] })
		pct := func(p float64) time.Duration {
			if len(lat) == 0 {
				return 0
			}
			return lat[min(len(lat)-1, int(float64(len(lat)-1)*p))]
		}
		fmt.Printf("total=%d success=%d failures=%d throughput=%.2f req/s p50=%s p95=%s p99=%s\n", total, ok.Load(), fail.Load(), float64(total)/time.Since(start).Seconds(), pct(.5), pct(.95), pct(.99))
		return nil
	}}
	c.Flags().StringVar(&target, "url", "", "target URL")
	c.Flags().IntVar(&total, "requests", 100, "request count")
	c.Flags().IntVar(&concurrency, "concurrency", 10, "parallel workers")
	c.Flags().DurationVar(&timeout, "timeout", 3*time.Second, "request timeout")
	_ = c.MarkFlagRequired("url")
	return c
}
