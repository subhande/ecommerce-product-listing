package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	productsPath = "/api/v1/products"
)

type priceRange struct {
	Min float64 `json:"min"`
	Max float64 `json:"max"`
}

type valuePool struct {
	Categories  []string     `json:"categories"`
	Brands      []string     `json:"brands"`
	Currencies  []string     `json:"currencies"`
	Countries   []string     `json:"countries"`
	SearchTerms []string     `json:"search_terms"`
	SortBy      []string     `json:"sort_by_columns"`
	SortOrders  []string     `json:"sort_orders"`
	SearchTypes []string     `json:"search_types"`
	PriceRanges []priceRange `json:"price_ranges"`
	Ratings     []float64    `json:"ratings"`
	ReviewCount []int        `json:"review_counts"`
	PageSizes   []int        `json:"page_sizes"`
}

type productPayload struct {
	Title             string  `json:"title"`
	ASIN              string  `json:"asin"`
	Description       string  `json:"description"`
	Category          string  `json:"category"`
	Brand             string  `json:"brand"`
	ImageURL          string  `json:"image_url"`
	ProductURL        string  `json:"product_url"`
	Price             float64 `json:"price"`
	Currency          string  `json:"currency"`
	Country           string  `json:"country"`
	Stock             int     `json:"stock"`
	AvgRating         float64 `json:"avg_rating"`
	ReviewCount       int     `json:"review_count"`
	BoughtInLastMonth int     `json:"bought_in_last_month"`
	IsBestSeller      bool    `json:"is_best_seller"`
}

type endpointType string

const (
	endpointGet  endpointType = "GET /api/v1/products"
	endpointPost endpointType = "POST /api/v1/products"
	endpointBulk endpointType = "POST /api/v1/products/bulk"
)

type runConfig struct {
	BaseURL       string        `json:"base_url"`
	TotalRequests int           `json:"total_requests"`
	Concurrency   int           `json:"concurrency"`
	BulkSize      int           `json:"bulk_size"`
	Timeout       time.Duration `json:"timeout"`
	ValuesFile    string        `json:"values_file"`
	OutputPath    string        `json:"output_path"`
	Seed          int64         `json:"seed"`
	GetWeight     int           `json:"get_weight"`
	PostWeight    int           `json:"post_weight"`
	BulkWeight    int           `json:"bulk_weight"`
}

type requestSample struct {
	Endpoint  endpointType
	Duration  time.Duration
	Status    int
	Success   bool
	ErrorText string
}

type endpointAccumulator struct {
	Endpoint     endpointType
	Requests     int
	Successes    int
	Failures     int
	StatusCodes  map[string]int
	LatenciesMS  []float64
	ErrorSamples []string
}

type latencyStats struct {
	Min float64 `json:"min"`
	Max float64 `json:"max"`
	Avg float64 `json:"avg"`
	P50 float64 `json:"p50"`
	P90 float64 `json:"p90"`
	P95 float64 `json:"p95"`
	P99 float64 `json:"p99"`
}

type endpointResult struct {
	Endpoint     endpointType `json:"endpoint"`
	Requests     int          `json:"requests"`
	Successes    int          `json:"successes"`
	Failures     int          `json:"failures"`
	SuccessRate  float64      `json:"success_rate"`
	StatusCodes  map[string]int
	LatencyMS    latencyStats `json:"latency_ms"`
	ErrorSamples []string     `json:"error_samples,omitempty"`
}

type loadTestReport struct {
	StartedAt     time.Time        `json:"started_at"`
	FinishedAt    time.Time        `json:"finished_at"`
	DurationMS    int64            `json:"duration_ms"`
	Config        runConfig        `json:"config"`
	Totals        endpointResult   `json:"totals"`
	EndpointStats []endpointResult `json:"endpoint_stats"`
}

func main() {
	cfg := parseFlags()

	if cfg.TotalRequests <= 0 || cfg.Concurrency <= 0 || cfg.BulkSize <= 0 {
		fmt.Fprintln(os.Stderr, "total-requests, concurrency, and bulk-size must be > 0")
		os.Exit(1)
	}
	if cfg.GetWeight < 0 || cfg.PostWeight < 0 || cfg.BulkWeight < 0 || cfg.GetWeight+cfg.PostWeight+cfg.BulkWeight == 0 {
		fmt.Fprintln(os.Stderr, "weights must be >= 0 and at least one weight must be > 0")
		os.Exit(1)
	}

	values, err := loadValuePool(cfg.ValuesFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed loading values file: %v\n", err)
		os.Exit(1)
	}

	client := &http.Client{Timeout: cfg.Timeout}
	acc := newAccumulator()
	start := time.Now()

	jobs := make(chan int)
	var wg sync.WaitGroup
	for i := 0; i < cfg.Concurrency; i++ {
		wg.Add(1)
		workerID := i
		go func() {
			defer wg.Done()
			r := rand.New(rand.NewSource(cfg.Seed + int64(workerID+1)))
			for range jobs {
				ep := pickEndpoint(r, cfg)
				sample := runScenario(client, r, strings.TrimRight(cfg.BaseURL, "/"), ep, cfg.BulkSize, values)
				acc.add(sample)
			}
		}()
	}

	for i := 0; i < cfg.TotalRequests; i++ {
		jobs <- i
	}
	close(jobs)
	wg.Wait()

	finish := time.Now()
	report := acc.report(start, finish, cfg)
	if err := writeReport(cfg.OutputPath, report); err != nil {
		fmt.Fprintf(os.Stderr, "failed writing output file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("load test completed: %d requests, concurrency=%d\n", cfg.TotalRequests, cfg.Concurrency)
	fmt.Printf("result file: %s\n", cfg.OutputPath)
}

func parseFlags() runConfig {
	cfg := runConfig{}
	flag.StringVar(&cfg.BaseURL, "base-url", "http://localhost:8080", "Base API URL")
	flag.IntVar(&cfg.TotalRequests, "total-requests", 1000, "Total requests across all endpoints")
	flag.IntVar(&cfg.Concurrency, "concurrency", 20, "Number of concurrent workers")
	flag.IntVar(&cfg.BulkSize, "bulk-size", 20, "Number of products per bulk request")
	flag.DurationVar(&cfg.Timeout, "timeout", 10*time.Second, "HTTP request timeout")
	flag.StringVar(&cfg.ValuesFile, "values-file", "loadtest/query_values.json", "JSON file for query/data values")
	flag.StringVar(&cfg.OutputPath, "output", "", "Output JSON result file path")
	flag.IntVar(&cfg.GetWeight, "get-weight", 50, "Weight for GET /api/v1/products")
	flag.IntVar(&cfg.PostWeight, "post-weight", 25, "Weight for POST /api/v1/products")
	flag.IntVar(&cfg.BulkWeight, "bulk-weight", 25, "Weight for POST /api/v1/products/bulk")
	flag.Int64Var(&cfg.Seed, "seed", time.Now().UnixNano(), "Random seed")
	flag.Parse()

	if cfg.OutputPath == "" {
		ts := time.Now().Format("20060102_150405")
		cfg.OutputPath = filepath.Join("results", fmt.Sprintf("load_test_%s.json", ts))
	}
	return cfg
}

func loadValuePool(path string) (valuePool, error) {
	pool := defaultValuePool()
	b, err := os.ReadFile(path)
	if err != nil {
		return pool, err
	}
	var parsed valuePool
	if err := json.Unmarshal(b, &parsed); err != nil {
		return pool, err
	}

	if len(parsed.Categories) > 0 {
		pool.Categories = parsed.Categories
	}
	if len(parsed.Brands) > 0 {
		pool.Brands = parsed.Brands
	}
	if len(parsed.Currencies) > 0 {
		pool.Currencies = parsed.Currencies
	}
	if len(parsed.Countries) > 0 {
		pool.Countries = parsed.Countries
	}
	if len(parsed.SearchTerms) > 0 {
		pool.SearchTerms = parsed.SearchTerms
	}
	if len(parsed.SortBy) > 0 {
		pool.SortBy = parsed.SortBy
	}
	if len(parsed.SortOrders) > 0 {
		pool.SortOrders = parsed.SortOrders
	}
	if len(parsed.SearchTypes) > 0 {
		pool.SearchTypes = parsed.SearchTypes
	}
	if len(parsed.PriceRanges) > 0 {
		pool.PriceRanges = parsed.PriceRanges
	}
	if len(parsed.Ratings) > 0 {
		pool.Ratings = parsed.Ratings
	}
	if len(parsed.ReviewCount) > 0 {
		pool.ReviewCount = parsed.ReviewCount
	}
	if len(parsed.PageSizes) > 0 {
		pool.PageSizes = parsed.PageSizes
	}
	return pool, nil
}

func defaultValuePool() valuePool {
	return valuePool{
		Categories:  []string{"Graphics Cards", "Laptops", "Headphones", "Smartphones"},
		Brands:      []string{"Samsung", "Apple", "Sony", "Dell", "ASUS"},
		Currencies:  []string{"USD", "EUR"},
		Countries:   []string{"US", "UK"},
		SearchTerms: []string{"speaker", "gaming", "wireless"},
		SortBy:      []string{"price", "bought_in_last_month", "avg_rating", "updated_at"},
		SortOrders:  []string{"asc", "desc"},
		SearchTypes: []string{"simple", "fts"},
		PriceRanges: []priceRange{{Min: 10, Max: 100}, {Min: 100, Max: 500}, {Min: 500, Max: 2000}},
		Ratings:     []float64{3.0, 3.5, 4.0, 4.5},
		ReviewCount: []int{10, 50, 100, 500},
		PageSizes:   []int{20, 50, 100},
	}
}

func pickEndpoint(r *rand.Rand, cfg runConfig) endpointType {
	total := cfg.GetWeight + cfg.PostWeight + cfg.BulkWeight
	n := r.Intn(total)
	if n < cfg.GetWeight {
		return endpointGet
	}
	n -= cfg.GetWeight
	if n < cfg.PostWeight {
		return endpointPost
	}
	return endpointBulk
}

func runScenario(client *http.Client, r *rand.Rand, baseURL string, endpoint endpointType, bulkSize int, pool valuePool) requestSample {
	switch endpoint {
	case endpointGet:
		return callGetProducts(client, r, baseURL, pool)
	case endpointPost:
		return callCreateProduct(client, r, baseURL, pool)
	case endpointBulk:
		return callCreateProductsBulk(client, r, baseURL, bulkSize, pool)
	default:
		return requestSample{Endpoint: endpoint, Success: false, ErrorText: "unknown endpoint"}
	}
}

func callGetProducts(client *http.Client, r *rand.Rand, baseURL string, pool valuePool) requestSample {
	params := randomQueryParams(r, pool)
	target := baseURL + productsPath + "?" + params.Encode()

	req, _ := http.NewRequest(http.MethodGet, target, nil)
	start := time.Now()
	resp, err := client.Do(req)
	return parseResponse(endpointGet, start, resp, err)
}

func callCreateProduct(client *http.Client, r *rand.Rand, baseURL string, pool valuePool) requestSample {
	product := randomProduct(r, pool)
	body, err := json.Marshal(product)
	if err != nil {
		return requestSample{Endpoint: endpointPost, Success: false, ErrorText: err.Error()}
	}

	req, _ := http.NewRequest(http.MethodPost, baseURL+productsPath, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	start := time.Now()
	resp, err := client.Do(req)
	return parseResponse(endpointPost, start, resp, err)
}

func callCreateProductsBulk(client *http.Client, r *rand.Rand, baseURL string, bulkSize int, pool valuePool) requestSample {
	products := make([]productPayload, 0, bulkSize)
	for i := 0; i < bulkSize; i++ {
		products = append(products, randomProduct(r, pool))
	}
	body, err := json.Marshal(products)
	if err != nil {
		return requestSample{Endpoint: endpointBulk, Success: false, ErrorText: err.Error()}
	}

	req, _ := http.NewRequest(http.MethodPost, baseURL+productsPath+"/bulk", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	start := time.Now()
	resp, err := client.Do(req)
	return parseResponse(endpointBulk, start, resp, err)
}

func parseResponse(ep endpointType, start time.Time, resp *http.Response, err error) requestSample {
	sample := requestSample{
		Endpoint: ep,
		Duration: time.Since(start),
	}
	if err != nil {
		sample.Success = false
		sample.ErrorText = err.Error()
		return sample
	}
	defer resp.Body.Close()

	sample.Status = resp.StatusCode
	sample.Success = resp.StatusCode >= 200 && resp.StatusCode < 300
	if !sample.Success {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		bodyMsg := strings.TrimSpace(string(b))
		if bodyMsg == "" {
			bodyMsg = resp.Status
		}
		sample.ErrorText = bodyMsg
	}
	return sample
}

func randomProduct(r *rand.Rand, pool valuePool) productPayload {
	adjectives := []string{"Ultra", "Pro", "Smart", "Elite", "Rapid", "Prime", "Core", "Flex"}
	items := []string{"Speaker", "Laptop", "Headset", "Keyboard", "Monitor", "SSD", "Phone", "Camera"}

	category := pickString(r, pool.Categories)
	brand := pickString(r, pool.Brands)
	item := pickString(r, items)
	title := fmt.Sprintf("%s %s %s", brand, pickString(r, adjectives), item)
	asin := randomASIN(r, 10)
	desc := fmt.Sprintf("%s designed for %s workloads and daily use.", title, strings.ToLower(category))
	price := float64(r.Intn(300000)+500) / 100.0
	slug := strings.ToLower(strings.ReplaceAll(title, " ", "-"))

	return productPayload{
		Title:             title,
		ASIN:              asin,
		Description:       desc,
		Category:          category,
		Brand:             brand,
		ImageURL:          "https://example.com/images/" + slug + ".jpg",
		ProductURL:        "https://example.com/products/" + slug + "-" + asin,
		Price:             price,
		Currency:          pickString(r, pool.Currencies),
		Country:           pickString(r, pool.Countries),
		Stock:             r.Intn(500),
		AvgRating:         float64(r.Intn(31)+20) / 10.0,
		ReviewCount:       r.Intn(5000),
		BoughtInLastMonth: r.Intn(10000),
		IsBestSeller:      r.Intn(100) < 20,
	}
}

func randomQueryParams(r *rand.Rand, pool valuePool) url.Values {
	v := url.Values{}
	v.Set("sort_by_column", pickString(r, pool.SortBy))
	v.Set("sort_order", pickString(r, pool.SortOrders))
	v.Set("page_size", strconv.Itoa(pickInt(r, pool.PageSizes)))

	kind := r.Intn(6)
	switch kind {
	case 0:
		// sort-only query
	case 1:
		v.Set("category", pickString(r, pool.Categories))
	case 2:
		v.Set("brand", pickString(r, pool.Brands))
		pr := pickRange(r, pool.PriceRanges)
		v.Set("min_price", fmt.Sprintf("%.2f", pr.Min))
		v.Set("max_price", fmt.Sprintf("%.2f", pr.Max))
	case 3:
		v.Set("search_query_text", pickString(r, pool.SearchTerms))
		v.Set("search_type", pickString(r, pool.SearchTypes))
	case 4:
		v.Set("rating_more_than_equal", fmt.Sprintf("%.1f", pickFloat(r, pool.Ratings)))
		v.Set("review_count", strconv.Itoa(pickInt(r, pool.ReviewCount)))
	default:
		v.Set("category", pickString(r, pool.Categories))
		v.Set("search_query_text", pickString(r, pool.SearchTerms))
		v.Set("search_type", pickString(r, pool.SearchTypes))
		v.Set("page_number", strconv.Itoa(r.Intn(20)+1))
	}

	if r.Intn(100) < 15 {
		v.Set("show_out_of_stock", "true")
	}
	return v
}

func randomASIN(r *rand.Rand, n int) string {
	const letters = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[r.Intn(len(letters))]
	}
	return string(b)
}

func pickString(r *rand.Rand, arr []string) string {
	if len(arr) == 0 {
		return ""
	}
	return arr[r.Intn(len(arr))]
}

func pickInt(r *rand.Rand, arr []int) int {
	if len(arr) == 0 {
		return 1
	}
	return arr[r.Intn(len(arr))]
}

func pickFloat(r *rand.Rand, arr []float64) float64 {
	if len(arr) == 0 {
		return 0
	}
	return arr[r.Intn(len(arr))]
}

func pickRange(r *rand.Rand, arr []priceRange) priceRange {
	if len(arr) == 0 {
		return priceRange{}
	}
	return arr[r.Intn(len(arr))]
}

type accumulator struct {
	mu   sync.Mutex
	data map[endpointType]*endpointAccumulator
}

func newAccumulator() *accumulator {
	return &accumulator{
		data: map[endpointType]*endpointAccumulator{
			endpointGet:  {Endpoint: endpointGet, StatusCodes: map[string]int{}},
			endpointPost: {Endpoint: endpointPost, StatusCodes: map[string]int{}},
			endpointBulk: {Endpoint: endpointBulk, StatusCodes: map[string]int{}},
		},
	}
}

func (a *accumulator) add(sample requestSample) {
	a.mu.Lock()
	defer a.mu.Unlock()

	row := a.data[sample.Endpoint]
	if row == nil {
		row = &endpointAccumulator{Endpoint: sample.Endpoint, StatusCodes: map[string]int{}}
		a.data[sample.Endpoint] = row
	}
	row.Requests++
	row.LatenciesMS = append(row.LatenciesMS, float64(sample.Duration.Microseconds())/1000.0)
	if sample.Success {
		row.Successes++
	} else {
		row.Failures++
		if sample.ErrorText != "" && len(row.ErrorSamples) < 5 {
			row.ErrorSamples = append(row.ErrorSamples, sample.ErrorText)
		}
	}
	if sample.Status > 0 {
		k := strconv.Itoa(sample.Status)
		row.StatusCodes[k]++
	} else if sample.ErrorText != "" {
		row.StatusCodes["transport_error"]++
	}
}

func (a *accumulator) report(start, finish time.Time, cfg runConfig) loadTestReport {
	a.mu.Lock()
	defer a.mu.Unlock()

	endpoints := make([]endpointResult, 0, len(a.data))
	total := endpointResult{
		Endpoint:    "ALL",
		StatusCodes: map[string]int{},
	}
	allLatencies := make([]float64, 0, cfg.TotalRequests)

	keys := []endpointType{endpointGet, endpointPost, endpointBulk}
	for _, key := range keys {
		row := a.data[key]
		if row == nil {
			continue
		}
		res := endpointResult{
			Endpoint:     row.Endpoint,
			Requests:     row.Requests,
			Successes:    row.Successes,
			Failures:     row.Failures,
			StatusCodes:  row.StatusCodes,
			LatencyMS:    computeLatencyStats(row.LatenciesMS),
			ErrorSamples: row.ErrorSamples,
		}
		if row.Requests > 0 {
			res.SuccessRate = float64(row.Successes) * 100.0 / float64(row.Requests)
		}
		endpoints = append(endpoints, res)

		total.Requests += row.Requests
		total.Successes += row.Successes
		total.Failures += row.Failures
		for k, v := range row.StatusCodes {
			total.StatusCodes[k] += v
		}
		allLatencies = append(allLatencies, row.LatenciesMS...)
	}
	if total.Requests > 0 {
		total.SuccessRate = float64(total.Successes) * 100.0 / float64(total.Requests)
	}
	total.LatencyMS = computeLatencyStats(allLatencies)

	return loadTestReport{
		StartedAt:     start,
		FinishedAt:    finish,
		DurationMS:    finish.Sub(start).Milliseconds(),
		Config:        cfg,
		Totals:        total,
		EndpointStats: endpoints,
	}
}

func computeLatencyStats(values []float64) latencyStats {
	if len(values) == 0 {
		return latencyStats{}
	}
	cp := append([]float64(nil), values...)
	sort.Float64s(cp)

	sum := 0.0
	for _, v := range cp {
		sum += v
	}
	return latencyStats{
		Min: cp[0],
		Max: cp[len(cp)-1],
		Avg: sum / float64(len(cp)),
		P50: percentile(cp, 50),
		P90: percentile(cp, 90),
		P95: percentile(cp, 95),
		P99: percentile(cp, 99),
	}
}

func percentile(sorted []float64, p float64) float64 {
	if len(sorted) == 0 {
		return 0
	}
	if p <= 0 {
		return sorted[0]
	}
	if p >= 100 {
		return sorted[len(sorted)-1]
	}
	pos := (p / 100.0) * float64(len(sorted)-1)
	lower := int(pos)
	upper := lower + 1
	if upper >= len(sorted) {
		return sorted[lower]
	}
	weight := pos - float64(lower)
	return sorted[lower]*(1-weight) + sorted[upper]*weight
}

func writeReport(path string, report loadTestReport) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	out, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, out, 0o644)
}
