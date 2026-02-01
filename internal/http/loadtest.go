package http

import (
	"fmt"
	"sort"
	"sync"
	"time"
)

// LoadTestConfig defines parameters for load testing
type LoadTestConfig struct {
	Request       Request       // The request to test
	Concurrency   int           // Number of concurrent workers
	TotalRequests int           // Total number of requests to send
	Duration      time.Duration // Alternative: run for specified duration (0 = use TotalRequests)
	RampUpTime    time.Duration // Time to gradually increase to full concurrency
}

// LoadTestResult contains aggregated results from load test
type LoadTestResult struct {
	TotalRequests      int
	SuccessfulRequests int
	FailedRequests     int
	TotalDuration      time.Duration

	// Response times
	MinResponseTime    time.Duration
	MaxResponseTime    time.Duration
	AvgResponseTime    time.Duration
	MedianResponseTime time.Duration
	P95ResponseTime    time.Duration
	P99ResponseTime    time.Duration

	// Status codes
	StatusCodes map[int]int

	// Errors
	Errors map[string]int

	// Throughput
	RequestsPerSecond float64

	// Individual results
	IndividualResults []LoadTestRequestResult
}

// LoadTestRequestResult contains result for a single request in load test
type LoadTestRequestResult struct {
	StatusCode   int
	ResponseTime time.Duration
	Error        error
	Timestamp    time.Time
}

// RunLoadTest executes a load test with the given configuration
func RunLoadTest(client *Client, config LoadTestConfig, progressCallback func(completed, total int)) (*LoadTestResult, error) {
	if config.Concurrency < 1 {
		config.Concurrency = 1
	}

	if config.TotalRequests < 1 && config.Duration == 0 {
		return nil, fmt.Errorf("either TotalRequests or Duration must be specified")
	}

	startTime := time.Now()

	results := &LoadTestResult{
		StatusCodes:       make(map[int]int),
		Errors:            make(map[string]int),
		IndividualResults: []LoadTestRequestResult{},
	}

	var mu sync.Mutex
	var wg sync.WaitGroup

	// Channel to control work distribution
	workChan := make(chan int, config.Concurrency)
	resultsChan := make(chan LoadTestRequestResult, config.Concurrency)

	// Collect results
	go func() {
		for result := range resultsChan {
			mu.Lock()
			results.IndividualResults = append(results.IndividualResults, result)
			results.TotalRequests++

			if result.Error != nil {
				results.FailedRequests++
				errMsg := result.Error.Error()
				results.Errors[errMsg]++
			} else {
				results.SuccessfulRequests++
				results.StatusCodes[result.StatusCode]++
			}

			if progressCallback != nil {
				progressCallback(results.TotalRequests, config.TotalRequests)
			}
			mu.Unlock()
		}
	}()

	// Start workers
	for i := 0; i < config.Concurrency; i++ {
		wg.Add(1)

		// Ramp-up delay
		if config.RampUpTime > 0 {
			delay := time.Duration(float64(config.RampUpTime) * float64(i) / float64(config.Concurrency))
			time.Sleep(delay)
		}

		go func() {
			defer wg.Done()

			for range workChan {
				requestStart := time.Now()
				response := client.Send(config.Request)
				requestEnd := time.Now()

				result := LoadTestRequestResult{
					StatusCode:   response.StatusCode,
					ResponseTime: requestEnd.Sub(requestStart),
					Error:        response.Error,
					Timestamp:    requestStart,
				}

				resultsChan <- result
			}
		}()
	}

	// Send work
	if config.Duration > 0 {
		// Duration-based load test
		stopTime := time.Now().Add(config.Duration)
		requestCount := 0
		for time.Now().Before(stopTime) {
			workChan <- requestCount
			requestCount++
		}
		config.TotalRequests = requestCount // Update for progress reporting
	} else {
		// Request count-based load test
		for i := 0; i < config.TotalRequests; i++ {
			workChan <- i
		}
	}

	close(workChan)
	wg.Wait()
	close(resultsChan)

	// Wait for all results to be collected
	time.Sleep(100 * time.Millisecond)

	endTime := time.Now()
	results.TotalDuration = endTime.Sub(startTime)

	// Calculate statistics
	calculateStatistics(results)

	return results, nil
}

// calculateStatistics computes aggregated statistics from individual results
func calculateStatistics(results *LoadTestResult) {
	if len(results.IndividualResults) == 0 {
		return
	}

	// Collect response times for successful requests
	var responseTimes []time.Duration
	var totalResponseTime time.Duration

	for _, result := range results.IndividualResults {
		if result.Error == nil {
			responseTimes = append(responseTimes, result.ResponseTime)
			totalResponseTime += result.ResponseTime
		}
	}

	if len(responseTimes) == 0 {
		return
	}

	// Sort for percentile calculation
	sort.Slice(responseTimes, func(i, j int) bool {
		return responseTimes[i] < responseTimes[j]
	})

	// Min and Max
	results.MinResponseTime = responseTimes[0]
	results.MaxResponseTime = responseTimes[len(responseTimes)-1]

	// Average
	results.AvgResponseTime = totalResponseTime / time.Duration(len(responseTimes))

	// Median (P50)
	medianIdx := len(responseTimes) / 2
	if len(responseTimes)%2 == 0 {
		results.MedianResponseTime = (responseTimes[medianIdx-1] + responseTimes[medianIdx]) / 2
	} else {
		results.MedianResponseTime = responseTimes[medianIdx]
	}

	// P95
	p95Idx := int(float64(len(responseTimes)) * 0.95)
	if p95Idx >= len(responseTimes) {
		p95Idx = len(responseTimes) - 1
	}
	results.P95ResponseTime = responseTimes[p95Idx]

	// P99
	p99Idx := int(float64(len(responseTimes)) * 0.99)
	if p99Idx >= len(responseTimes) {
		p99Idx = len(responseTimes) - 1
	}
	results.P99ResponseTime = responseTimes[p99Idx]

	// Requests per second
	if results.TotalDuration.Seconds() > 0 {
		results.RequestsPerSecond = float64(results.TotalRequests) / results.TotalDuration.Seconds()
	}
}

// FormatLoadTestResult returns a human-readable summary of load test results
func FormatLoadTestResult(result *LoadTestResult) string {
	output := fmt.Sprintf("Load Test Results\n")
	output += fmt.Sprintf("==================\n\n")

	output += "Summary:\n"
	output += fmt.Sprintf("  Total Requests:      %d\n", result.TotalRequests)
	output += fmt.Sprintf("  Successful:          %d (%.1f%%)\n", result.SuccessfulRequests,
		float64(result.SuccessfulRequests)/float64(result.TotalRequests)*100)
	output += fmt.Sprintf("  Failed:              %d (%.1f%%)\n", result.FailedRequests,
		float64(result.FailedRequests)/float64(result.TotalRequests)*100)
	output += fmt.Sprintf("  Total Duration:      %v\n", result.TotalDuration)
	output += fmt.Sprintf("  Requests/sec:        %.2f\n\n", result.RequestsPerSecond)

	if result.SuccessfulRequests > 0 {
		output += fmt.Sprintf("Response Times:\n")
		output += fmt.Sprintf("  Min:                 %v\n", result.MinResponseTime)
		output += fmt.Sprintf("  Max:                 %v\n", result.MaxResponseTime)
		output += fmt.Sprintf("  Avg:                 %v\n", result.AvgResponseTime)
		output += fmt.Sprintf("  Median (P50):        %v\n", result.MedianResponseTime)
		output += fmt.Sprintf("  95th Percentile:     %v\n", result.P95ResponseTime)
		output += fmt.Sprintf("  99th Percentile:     %v\n\n", result.P99ResponseTime)
	}

	if len(result.StatusCodes) > 0 {
		output += fmt.Sprintf("Status Codes:\n")
		// Sort status codes for consistent output
		var codes []int
		for code := range result.StatusCodes {
			codes = append(codes, code)
		}
		sort.Ints(codes)
		for _, code := range codes {
			count := result.StatusCodes[code]
			output += fmt.Sprintf("  %d:                   %d (%.1f%%)\n", code, count,
				float64(count)/float64(result.TotalRequests)*100)
		}
		output += "\n"
	}

	if len(result.Errors) > 0 {
		output += fmt.Sprintf("Errors:\n")
		for errMsg, count := range result.Errors {
			output += fmt.Sprintf("  %s: %d\n", errMsg, count)
		}
	}

	return output
}
