package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math"
	"math/rand"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Babatunde13/event-pipeline/internal/event"
	"github.com/joho/godotenv"
)

type job struct {
	url string
	evt event.Event
}

func getBaseURL() string {
	godotenv.Load()
	if base := os.Getenv("BASE_URL"); base == "" {
		panic("BASE_URL environment variable is not set, please set it to the base URL of your API")
	} else {
		return base
	}
}

func buildHTTPClient(maxConns int, timeout time.Duration) *http.Client {
	tr := &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		DialContext:           (&net.Dialer{Timeout: 3 * time.Second, KeepAlive: 30 * time.Second}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          maxConns * 2,
		MaxIdleConnsPerHost:   maxConns,
		MaxConnsPerHost:       maxConns, // cap to avoid unbounded growth
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   3 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
	return &http.Client{
		Transport: tr,
		Timeout:   timeout,
	}
}

func sendEvent(client *http.Client, targetURL string, evt event.Event) (int, time.Duration, error) {
	body, err := json.Marshal(evt)
	if err != nil {
		return 0, 0, fmt.Errorf("marshal: %w", err)
	}
	start := time.Now()
	resp, err := client.Post(targetURL, "application/json", bytes.NewReader(body))
	lat := time.Since(start)
	if err != nil {
		return 0, lat, err
	}
	defer resp.Body.Close()
	return resp.StatusCode, lat, nil
}

func main() {
	var (
		eps         = flag.Int("eps", 500, "Total events per second to generate across selected targets")
		durationStr = flag.String("duration", "15m", "Test duration, for example 15m")
		targetType  = flag.String("type", "", "Target: kafka, eventbridge, or empty for both")
		concurrency = flag.Int("concurrency", 200, "Number of concurrent workers")
		timeoutMs   = flag.Int("timeout_ms", 3000, "Per request timeout in milliseconds")
		rampStr     = flag.String("ramp", "0s", "Optional linear ramp up duration, e.g. 30s")
	)
	flag.Parse()
	base := getBaseURL()
	fmt.Println("Base URL:", base)

	dur, err := time.ParseDuration(*durationStr)
	if err != nil {
		log.Fatalf("invalid duration: %v", err)
	}
	ramp, err := time.ParseDuration(*rampStr)
	if err != nil {
		log.Fatalf("invalid ramp: %v", err)
	}
	targets := map[string]string{
		"kafka":       base + "/kafka",
		"eventbridge": base + "/eventbridge",
	}

	selected := []string{}
	switch strings.ToLower(*targetType) {
	case "":
		selected = []string{"kafka", "eventbridge"}
	case "kafka", "eventbridge":
		selected = []string{strings.ToLower(*targetType)}
	default:
		log.Fatalf("invalid -type: %s", *targetType)
	}
	// Per target EPS so that the total is exactly -eps
	perTargetEPS := float64(*eps) / float64(len(selected))

	log.Printf("Starting load: total EPS=%d, per-target EPS=%.2f, duration=%s, concurrency=%d, ramp=%s",
		*eps, perTargetEPS, dur, *concurrency, ramp)

	// Data generators
	rand.New(rand.NewSource(time.Now().UnixNano())) // Seed for random number generation
	eventTypes := []string{"view_product", "add_to_cart", "checkout"}
	users := []string{"user1", "user2", "user3", "user4", "user5"}
	products := []string{"prod1", "prod2", "prod3", "prod4"}

	buildEvent := func() event.Event {
		return event.New(
			event.EventTypes[rand.Intn(len(eventTypes))],
			users[rand.Intn(len(users))],
			map[string]interface{}{
				"product_id": products[rand.Intn(len(products))],
				"price":      math.Round((rand.Float64()*100)*100) / 100, // 2dp
			},
		)
	}

	// Worker pool
	client := buildHTTPClient(*concurrency, time.Duration(*timeoutMs)*time.Millisecond)
	jobs := make(chan job, *eps*2) // small burst buffer
	ctx, cancel := context.WithTimeout(context.Background(), dur)
	defer cancel()

	var sent int64
	var okCount int64
	var errCount int64
	var bytesSent int64
	var latSumMicros int64

	wg := &sync.WaitGroup{}
	worker := func() {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			case j := <-jobs:
				status, lat, err := sendEvent(client, j.url, j.evt)
				atomic.AddInt64(&sent, 1)
				atomic.AddInt64(&latSumMicros, lat.Microseconds())
				if err != nil || status >= 400 {
					atomic.AddInt64(&errCount, 1)
				} else {
					atomic.AddInt64(&okCount, 1)
					atomic.AddInt64(&bytesSent, int64(len(j.evt.EventID))) // cheap placeholder
				}
			}
		}
	}

	for i := 0; i < *concurrency; i++ {
		wg.Add(1)
		go worker()
	}

	// Scheduler: 100 Hz tick. Each tick enqueues a slice of events for each target.
	// We keep fractional carry so the average per second is exact.
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	startWall := time.Now()
	lastReport := startWall
	var carry map[string]float64 = map[string]float64{}
	for _, t := range selected {
		carry[t] = 0
	}

	rampEnd := startWall.Add(ramp)

loop:
	for {
		select {
		case <-ctx.Done():
			break loop
		case now := <-ticker.C:
			elapsed := now.Sub(startWall)

			// Compute current allowed EPS under ramp
			rampFactor := 1.0
			if ramp > 0 && now.Before(rampEnd) {
				rampFactor = float64(now.Sub(startWall)) / float64(ramp)
				if rampFactor < 0 {
					rampFactor = 0
				}
				if rampFactor > 1 {
					rampFactor = 1
				}
			}

			for _, name := range selected {
				targetURL := targets[name]

				targetEPS := perTargetEPS * rampFactor
				perTick := targetEPS / 100.0 // 100 ticks per second
				carry[name] += perTick

				n := int(carry[name])
				if n > 0 {
					carry[name] -= float64(n)
					for i := 0; i < n; i++ {
						select {
						case <-ctx.Done():
							break loop
						case jobs <- job{url: targetURL, evt: buildEvent()}:
						}
					}
				}
			}

			// Lightweight periodic report each second
			if now.Sub(lastReport) >= time.Second {
				s := atomic.LoadInt64(&sent)
				ok := atomic.LoadInt64(&okCount)
				errs := atomic.LoadInt64(&errCount)
				latAvg := time.Duration(0)
				if s > 0 {
					latAvg = time.Duration(atomic.LoadInt64(&latSumMicros)/s) * time.Microsecond
				}
				achieved := float64(s) / time.Since(startWall).Seconds()
				log.Printf("t=%4.1fs sent=%d ok=%d errs=%d eps=%.1f avg_lat=%s",
					elapsed.Seconds(), s, ok, errs, achieved, latAvg)
				lastReport = now
			}
		}
	}

	// Drain and shutdown
	close(jobs)
	wg.Wait()

	total := atomic.LoadInt64(&sent)
	ok := atomic.LoadInt64(&okCount)
	errs := atomic.LoadInt64(&errCount)
	avgLat := time.Duration(0)
	if total > 0 {
		avgLat = time.Duration(atomic.LoadInt64(&latSumMicros)/total) * time.Microsecond
	}
	achieved := float64(total) / time.Since(startWall).Seconds()

	fmt.Println("====== Summary ======")
	fmt.Printf("Duration:        %s\n", time.Since(startWall).Truncate(time.Millisecond))
	fmt.Printf("Configured EPS:  %d (total)\n", *eps)
	fmt.Printf("Achieved EPS:    %.2f\n", achieved)
	fmt.Printf("Total sent:      %d\n", total)
	fmt.Printf("Success:         %d\n", ok)
	fmt.Printf("Errors:          %d\n", errs)
	fmt.Printf("Avg latency:     %s\n", avgLat)
	fmt.Printf("Concurrency:     %d workers\n", *concurrency)
}
