import json
import os


paths = [
    "results/response_times_10k.json",
    "results/response_times_100k.json",
    "results/response_times_1M.json",
]


data_10k = []
with open(paths[0], "r") as f:
    data_10k = json.load(f)

data_100k = []
with open(paths[1], "r") as f:
    data_100k = json.load(f)

data_1M = []
with open(paths[2], "r") as f:
    data_1M = json.load(f)

data = [
    (data_10k, 10000, "~10k"),
    (data_100k, 100000, "~100k"),
    (data_1M, 1000000, "~1M"),
]

keys = list(data_10k.keys())

rows = []

for key in keys:
    row_name = key.split("(")[0].strip()
    pagination_type = key.split("(")[1].split(")")[0].strip().split(" ")[0].strip()

    for d in data:
        metrics = d[0][key]
        length = d[1]
        length_in_words = d[2]
        metrics_length = len(metrics)
        page_nos = [m["page"] for m in d[0][key]]
        metrics = [m["response_time_ms"] for m in metrics]

        max_page_no = max(page_nos)
        # First Page, Last Page, Avg Response Time, Median Response Time, P90, P95, P99
        first_page = metrics[0]
        last_page = metrics[-1]
        avg_response_time = sum(metrics) / metrics_length
        median_response_time = metrics[metrics_length // 2]
        p90_response_time = metrics[int(metrics_length * 0.9)]
        p95_response_time = metrics[int(metrics_length * 0.95)]
        p99_response_time = metrics[int(metrics_length * 0.99)]

        rows.append(
            {
                "Query": row_name,
                "Type": pagination_type,
                "Datset Size": length_in_words,
                "Total Pages": max_page_no,
                "First Page": first_page,
                "Last Page": last_page,
                "Avg Response Time": avg_response_time,
                "Median Response Time": median_response_time,
                "P90 Response Time": p90_response_time,
                "P95 Response Time": p95_response_time,
                "P99 Response Time": p99_response_time,
            }
        )


# Convert to markdown table
markdown_table = "| Query | Type | Dataset Size | Total Pages | First Page | Last Page | Avg Response Time (ms) | Median Response Time (ms) | P90 Response Time (ms) | P95 Response Time (ms) | P99 Response Time (ms) |\n"
markdown_table += "|-------|------|--------------|-------------|------------|-----------|-----------------------|--------------------------|------------------------|------------------------|------------------------|\n"

for row in rows:
    markdown_table += f"| {row['Query']} | {row['Type']} | {row['Datset Size']} | {row['Total Pages']} | {row['First Page']} | {row['Last Page']} | {row['Avg Response Time']:.2f} | {row['Median Response Time']} | {row['P90 Response Time']} | {row['P95 Response Time']} | {row['P99 Response Time']} |\n"

print(markdown_table)
