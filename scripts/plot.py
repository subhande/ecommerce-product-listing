import json
import time, copy
import re, os

import requests
import matplotlib.pyplot as plt
import seaborn as sns
from fetch_products import safe_slug, queries
from pprint import pprint
import numpy as np

BASE_URL = "http://localhost:8080/api/v1/products"


FETCH_API = f"{BASE_URL}"


dataset_lens = ["10k", "100k", "1M", "1M+"]

PATHS = [
    f"results/response_times_10k.json",
    f"results/response_times_100k.json",
    f"results/response_times_1M.json",
    # f"results/response_times_1M+.json",
]

data = {}
for path in PATHS:
    if os.path.exists(path):
        with open(path, "r") as f:
            response_times = json.load(f)
            data[path.split(".")[0].split("_")[-1]] = response_times


def smooth_series(values, alpha=0.2):
    if not values:
        return values
    smoothed = [float(values[0])]
    for value in values[1:]:
        smoothed.append(alpha * float(value) + (1 - alpha) * smoothed[-1])
    return smoothed


def normalize_progress_axis(values):
    n = len(values)
    if n <= 1:
        return [0.0] * n
    return list(np.linspace(0, 100, n))


def adaptive_alpha(length):
    if length < 10:
        return 0.35
    if length < 100:
        return 0.25
    return 0.12


def downsample_points(x_vals, y_vals, max_points=1500):
    if len(x_vals) <= max_points:
        return x_vals, y_vals
    idx = np.linspace(0, len(x_vals) - 1, max_points).astype(int)
    return [x_vals[i] for i in idx], [y_vals[i] for i in idx]


# Plot graphs for each query
print("\n\nGenerating plots...")

# Set style for better-looking plots
sns.set_style("whitegrid")
plt.rcParams["figure.figsize"] = (10, 6)

# for name, times in response_times.items():
#     if not times:
#         print(f"No data for {name}, skipping...")
#         continue

#     # Extract page numbers and response times
#     pages = [item["page"] for item in times]
#     response_ms = [item["response_time_ms"] for item in times]
#     smooth_response_ms = smooth_series(response_ms)

#     # Create the plot
#     plt.figure(figsize=(10, 6))
#     plt.plot(pages, smooth_response_ms, linewidth=2.5, color="#2E86AB")
#     plt.fill_between(pages, smooth_response_ms, alpha=0.2, color="#2E86AB")

#     plt.xlabel("Page Number", fontsize=12, fontweight="bold")
#     plt.ylabel("Response Time (ms)", fontsize=12, fontweight="bold")
#     plt.title(f"Response Time vs Page Number - {name}", fontsize=14, fontweight="bold", pad=20)
#     plt.grid(True, alpha=0.3)

#     # # Add value labels on points
#     # for i, (page, rt) in enumerate(zip(pages, response_ms)):
#     #     plt.annotate(
#     #         f"{rt}ms",
#     #         xy=(page, rt),
#     #         xytext=(0, 10),
#     #         textcoords="offset points",
#     #         ha="center",
#     #         fontsize=9,
#     #         bbox=dict(boxstyle="round,pad=0.3", facecolor="yellow", alpha=0.5),
#     #     )

#     plt.tight_layout()

#     # Save the plot
#     filename = f"results/response_time_{safe_slug(name)}.png"
#     plt.savefig(filename, dpi=300, bbox_inches="tight")
#     print(f"Saved plot: {filename}")
#     plt.close()

# Plot each group of queries together for comparison
groups = {}
for query in queries:
    for size, d in data.items():
        group_name = query.get("group", "Other Queries")
        if group_name not in groups:
            groups[group_name] = []
        groups[group_name].append([query["name"], size])

pprint(groups)

for group_name, query_names in groups.items():
    plt.figure(figsize=(14, 8))
    colors = ["#2E86AB", "#A23B72", "#F18F01", "#C73E1D", "#6A994E"]
    group_series = []

    for idx, name in enumerate(query_names):
        name, size = name
        times = data.get(size, {}).get(name, [])
        name = f"{name} ({size})"
        if not times:
            continue

        points = [(item.get("page"), item.get("response_time_ms")) for item in times]
        points = [(p, rt) for p, rt in points if isinstance(rt, (int, float)) and rt > 0]
        if not points:
            continue
        points.sort(key=lambda x: x[0] if isinstance(x[0], (int, float)) else 0)

        response_ms = [rt for _, rt in points]
        x_progress = normalize_progress_axis(response_ms)
        smooth_response_ms = smooth_series(response_ms, alpha=adaptive_alpha(len(response_ms)))
        x_progress, smooth_response_ms = downsample_points(x_progress, smooth_response_ms)
        group_series.extend(smooth_response_ms)
        color = colors[idx % len(colors)]
        plt.plot(
            x_progress,
            smooth_response_ms,
            linewidth=2,
            label=name,
            color=color,
        )

    if group_series:
        min_val = max(min(group_series), 1e-3)
        max_val = max(group_series)
        if max_val / min_val >= 40:
            plt.yscale("log")

    plt.xlabel("Progress Through Results (%)", fontsize=12, fontweight="bold")
    plt.ylabel("Response Time (ms)", fontsize=12, fontweight="bold")
    plt.title(f"Response Time Comparison - {group_name}", fontsize=14, fontweight="bold", pad=20)
    plt.legend(loc="best", fontsize=9, ncol=2)
    plt.grid(True, alpha=0.3)
    plt.tight_layout()

    filename = f"results/response_time_comparison_{safe_slug(group_name)}.png"
    plt.savefig(filename, dpi=300, bbox_inches="tight")
    print(f"Saved plot: {filename}")
    plt.close()

# # Create a combined plot with all queries
# if response_times:
#     plt.figure(figsize=(14, 8))

#     colors = ["#2E86AB", "#A23B72", "#F18F01", "#C73E1D", "#6A994E"]

#     for idx, (name, times) in enumerate(response_times.items()):
#         if not times:
#             continue
#         pages = [item["page"] for item in times]
#         response_ms = [item["response_time_ms"] for item in times]
#         smooth_response_ms = smooth_series(response_ms)
#         color = colors[idx % len(colors)]
#         plt.plot(
#             pages,
#             smooth_response_ms,
#             linewidth=2,
#             label=name,
#             color=color,
#         )

#     plt.xlabel("Page Number", fontsize=12, fontweight="bold")
#     plt.ylabel("Response Time (ms)", fontsize=12, fontweight="bold")
#     plt.title("Response Time Comparison - All Queries", fontsize=14, fontweight="bold", pad=20)
#     plt.legend(loc="best", fontsize=10)
#     plt.grid(True, alpha=0.3)
#     plt.tight_layout()

#     plt.savefig("results/response_time_all_queries.png", dpi=300, bbox_inches="tight")
#     print(f"Saved combined plot: results/response_time_all_queries.png")
#     plt.close()

# print("\n✓ All plots generated successfully!")
