package main

import (
	"context"
	"ecommerce_product_listing/config"
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

type csvProduct struct {
	title             string
	asin              string
	description       string
	category          string
	brand             string
	imageURL          string
	productURL        string
	price             float64
	currency          string
	country           string
	stock             int
	avgRating         float64
	reviewCount       int
	boughtInLastMonth int
	isBestSeller      bool
}

func main() {
	csvPath := flag.String("csv", "external-data/processed_amazon_uk_products.csv", "path to CSV file")
	batchSize := flag.Int("batch-size", 10000, "number of rows per DB batch")
	limit := flag.Int("limit", -1, "max rows to insert; -1 inserts all rows")
	flag.Parse()

	if *batchSize <= 0 {
		log.Fatal("batch-size must be > 0")
	}

	config.LoadEnv()
	config.ConnectDB()
	defer config.DB.Close()

	inserted, err := loadCSVAndInsert(context.Background(), *csvPath, *batchSize, *limit)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Insert completed. Total rows inserted: %d", inserted)
}

func loadCSVAndInsert(ctx context.Context, path string, batchSize, limit int) (int, error) {
	file, err := os.Open(path)
	if err != nil {
		return 0, fmt.Errorf("open csv: %w", err)
	}
	defer file.Close()

	r := csv.NewReader(file)
	r.FieldsPerRecord = -1
	r.LazyQuotes = true

	headers, err := r.Read()
	if err != nil {
		return 0, fmt.Errorf("read csv headers: %w", err)
	}

	hdrIdx := make(map[string]int, len(headers))
	for i, h := range headers {
		hdrIdx[strings.TrimSpace(strings.ToLower(h))] = i
	}

	required := []string{"title", "price", "currency", "stock"}
	for _, col := range required {
		if _, ok := hdrIdx[col]; !ok {
			return 0, fmt.Errorf("required column %q not found in csv", col)
		}
	}

	rows := make([][]interface{}, 0, batchSize)
	totalInserted := 0
	totalRead := 0
	start := time.Now()

	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("Skipping malformed row %d: %v", totalRead+1, err)
			continue
		}

		totalRead++
		if limit != -1 && totalRead > limit {
			break
		}

		p := parseProduct(record, hdrIdx)
		rows = append(rows, []interface{}{
			p.title,
			p.asin,
			p.description,
			p.category,
			p.brand,
			p.imageURL,
			p.productURL,
			p.price,
			p.currency,
			p.country,
			p.stock,
			p.avgRating,
			p.reviewCount,
			p.boughtInLastMonth,
			p.isBestSeller,
		})

		if len(rows) >= batchSize {
			n, err := copyBatch(ctx, rows)
			if err != nil {
				return totalInserted, err
			}
			totalInserted += int(n)
			rows = rows[:0]
			log.Printf("Inserted %d rows so far...", totalInserted)
		}
	}

	if len(rows) > 0 {
		n, err := copyBatch(ctx, rows)
		if err != nil {
			return totalInserted, err
		}
		totalInserted += int(n)
	}

	log.Printf("Processed %d rows in %s", totalInserted, time.Since(start).Round(time.Millisecond))
	return totalInserted, nil
}

func copyBatch(ctx context.Context, rows [][]interface{}) (int64, error) {
	n, err := config.DB.CopyFrom(
		ctx,
		pgx.Identifier{"products"},
		[]string{
			"title",
			"asin",
			"description",
			"category",
			"brand",
			"image_url",
			"product_url",
			"price",
			"currency",
			"country",
			"stock",
			"avg_rating",
			"review_count",
			"bought_in_last_month",
			"is_best_seller",
		},
		pgx.CopyFromRows(rows),
	)
	if err != nil {
		return 0, fmt.Errorf("copy batch failed: %w", err)
	}
	return n, nil
}

func parseProduct(record []string, idx map[string]int) csvProduct {
	get := func(col string) string {
		i, ok := idx[col]
		if !ok || i >= len(record) {
			return ""
		}
		return strings.TrimSpace(record[i])
	}

	return csvProduct{
		title:             get("title"),
		asin:              get("asin"),
		description:       get("description"),
		category:          get("category"),
		brand:             get("brand"),
		imageURL:          get("image_url"),
		productURL:        get("product_url"),
		price:             parseFloat(get("price")),
		currency:          getDefault(get("currency"), "GBP"),
		country:           get("country"),
		stock:             parseInt(get("stock")),
		avgRating:         parseFloat(get("avg_rating")),
		reviewCount:       parseInt(get("review_count")),
		boughtInLastMonth: parseInt(get("bought_in_last_month")),
		isBestSeller:      parseBool(get("is_best_seller")),
	}
}

func parseFloat(s string) float64 {
	if s == "" {
		return 0
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return v
}

func parseInt(s string) int {
	if s == "" {
		return 0
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		f, ferr := strconv.ParseFloat(s, 64)
		if ferr != nil {
			return 0
		}
		return int(f)
	}
	return v
}

func parseBool(s string) bool {
	if s == "" {
		return false
	}
	v, err := strconv.ParseBool(strings.ToLower(s))
	if err == nil {
		return v
	}
	return s == "1" || strings.EqualFold(s, "yes")
}

func getDefault(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}
