go run ./cmd/loadtest \
  --base-url http://localhost:8080 \
  --total-requests 100000 \
  --concurrency 100 \
  --bulk-size 25 \
  --values-file loadtest/query_values.json \
  --output results/load_test_run.json