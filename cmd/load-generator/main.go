package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/Babatunde13/event-pipeline/internal/event"
	"github.com/joho/godotenv"
)

func getBaseURL() string {
	godotenv.Load() // Load environment variables from .env file
	baseUrl := os.Getenv("BASE_URL")
	if baseUrl == "" {
		log.Fatal("BASE_URL environment variable is not set")
	}
	return strings.TrimSuffix(baseUrl, "/") // Ensure no trailing slash
}

func sendEvent(wg *sync.WaitGroup, targetURL string, evt event.Event) {
	defer wg.Done()

	jsonData, err := json.Marshal(evt)
	if err != nil {
		log.Printf("[ERROR] Failed to serialize event: %v", err)
		return
	}

	_, err = http.Post(targetURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("[ERROR] Failed to send event: %v", err)
	}
}

func main() {
	count := flag.Int("count", 1000, "Number of events to generate")
	rate := flag.Int("rps", 10, "Requests per second")
	targetType := flag.String("type", "", "Target producer type: kafka, eventbridge or empty for both")
	flag.Parse()

	fmt.Println("count =", *count, " rate =", *rate, " targetType =", *targetType)
	targets := map[string]string{
		"kafka":       fmt.Sprintf("%s/kafka", getBaseURL()),
		"eventbridge": fmt.Sprintf("%s/eventbridge", getBaseURL()),
	}

	selectedTargets := []string{}
	if *targetType == "" {
		selectedTargets = []string{"kafka", "eventbridge"}
	} else if _, ok := targets[strings.ToLower(*targetType)]; ok {
		selectedTargets = []string{strings.ToLower(*targetType)}
	} else {
		log.Fatalf("Invalid type: %s. Must be one of: kafka, eventbridge", *targetType)
	}

	eventTypes := []string{"view_product", "add_to_cart", "checkout"}
	users := []string{"user1", "user2", "user3", "user4", "user5"}
	products := []string{"prod1", "prod2", "prod3", "prod4"}
	rand.New(rand.NewSource(time.Now().UnixNano()))
	delay := time.Second / time.Duration(*rate)
	start := time.Now()

	var wg sync.WaitGroup

	for i := 0; i < *count; i++ {
		for _, key := range selectedTargets {
			evt := event.New(
				event.EventTypes[rand.Intn(len(eventTypes))],
				users[rand.Intn(len(users))],
				map[string]interface{}{
					"product_id": products[rand.Intn(len(products))],
					"price":      rand.Float64() * 100,
				},
			)
			wg.Add(1)
			go sendEvent(&wg, targets[key], evt)
		}
		time.Sleep(delay)
	}

	wg.Wait()
	elapsed := time.Since(start)
	fmt.Printf("Sent %d events in %s\n", *count*len(selectedTargets), elapsed)
	fmt.Printf("Average time per event: %s\n", elapsed/time.Duration(*count*len(selectedTargets)))
	fmt.Printf("Events per second: %.2f\n", float64(*count*len(selectedTargets))/elapsed.Seconds())
	fmt.Printf("Total time: %s\n", elapsed)
	fmt.Printf("Total events: %d\n", *count*len(selectedTargets))
}
