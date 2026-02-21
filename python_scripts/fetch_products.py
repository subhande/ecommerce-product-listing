import json
import time, copy
import re

import requests
import matplotlib.pyplot as plt
import seaborn as sns
import pickle, os, time
from tqdm import tqdm

from insert import load_csv_and_insert_bulk

BASE_URL = "http://localhost:8080/api/v1/products"


FETCH_API = f"{BASE_URL}"

COUNT_API = f"{BASE_URL}/counts"

# dataset_size = 10000
# dataset_len = "10k"

# dataset_size = 100000
# dataset_len = "100k"

dataset_size = 1000000
dataset_len = "1M"


PATH = f"results/response_times_{dataset_len}.json"

if os.path.exists(PATH):
    with open(PATH, "r") as f:
        response_times = json.load(f)
else:
    response_times = {}

queries = [
    # Global Sorting
    {
        "name": "Popular Products (Keyset Pagination)",
        "group": "Popular Products",
        "size": dataset_len,
        "query": {"sort_by_column": "bought_in_last_month", "sort_order": "desc", "page_size": 50},
    },
    {
        "name": "Popular Products (Offset Pagination)",
        "group": "Popular Products",
        "size": dataset_len,
        "query": {"sort_by_column": "bought_in_last_month", "sort_order": "desc", "page_size": 50, "page_number": 1},
    },
    {
        "name": "Price Low to High (Keyset Pagination)",
        "group": "Price Sorting",
        "size": dataset_len,
        "query": {"sort_by_column": "price", "sort_order": "asc", "page_size": 50},
    },
    {
        "name": "Price Low to High (Offset Pagination)",
        "group": "Price Sorting",
        "size": dataset_len,
        "query": {"sort_by_column": "price", "sort_order": "asc", "page_size": 50, "page_number": 1},
    },
    {
        "name": "Rating High to Low (Keyset Pagination)",
        "group": "Rating Sorting",
        "size": dataset_len,
        "query": {"sort_by_column": "avg_rating", "sort_order": "desc", "page_size": 50},
    },
    {
        "name": "Rating High to Low (Offset Pagination)",
        "group": "Rating Sorting",
        "size": dataset_len,
        "query": {"sort_by_column": "avg_rating", "sort_order": "desc", "page_size": 50, "page_number": 1},
    },
    {
        "name": "Recently Updated (Keyset Pagination)",
        "group": "Recent Updates",
        "size": dataset_len,
        "query": {"sort_by_column": "updated_at", "sort_order": "desc", "page_size": 50},
    },
    {
        "name": "Recently Updated (Offset Pagination)",
        "group": "Recent Updates",
        "size": dataset_len,
        "query": {"sort_by_column": "updated_at", "sort_order": "desc", "page_size": 50, "page_number": 1},
    },
    ## Category + Low to How
    {
        "name": "Filter by Category - Price Low to High (Keyset Pagination)",
        "group": "Category + Price Sorting",
        "size": dataset_len,
        "query": {
            "category": "Graphics Cards",
            "sort_by_column": "price",
            "sort_order": "asc",
            "page_size": 50,
        },
    },
    {
        "name": "Filter by Category - Price Low to High (Offset Pagination)",
        "group": "Category + Price Sorting",
        "size": dataset_len,
        "query": {
            "category": "Graphics Cards",
            "sort_by_column": "price",
            "sort_order": "asc",
            "page_size": 50,
            "page_number": 1,
        },
    },
    ## Category + Popularity
    {
        "name": "Filter by Category - Popularity (Keyset Pagination)",
        "group": "Category + Popularity",
        "size": dataset_len,
        "query": {
            "category": "Graphics Cards",
            "sort_by_column": "bought_in_last_month",
            "sort_order": "desc",
            "page_size": 50,
        },
    },
    {
        "name": "Filter by Category - Popularity (Offset Pagination)",
        "group": "Category + Popularity",
        "size": dataset_len,
        "query": {
            "category": "Graphics Cards",
            "sort_by_column": "bought_in_last_month",
            "sort_order": "desc",
            "page_size": 50,
            "page_number": 1,
        },
    },
    ## Category + Newest
    {
        "name": "Filter by Category - Newest (Keyset Pagination)",
        "group": "Category + Newest",
        "size": dataset_len,
        "query": {
            "category": "Graphics Cards",
            "sort_by_column": "updated_at",
            "sort_order": "desc",
            "page_size": 50,
        },
    },
    {
        "name": "Filter by Category - Newest (Offset Pagination)",
        "group": "Category + Newest",
        "size": dataset_len,
        "query": {
            "category": "Graphics Cards",
            "sort_by_column": "updated_at",
            "sort_order": "desc",
            "page_size": 50,
            "page_number": 1,
        },
    },
    ## Search + Popularity - Simple Search with GIN Index
    {
        "name": "Search + Popularity - Simple Search with GIN Index (Keyset Pagination)",
        "group": "Search + Popularity",
        "size": dataset_len,
        "query": {
            "search_query_text": "speaker",
            "search_type": "simple",
            "sort_by_column": "bought_in_last_month",
            "sort_order": "desc",
            "page_size": 50,
            "size": dataset_len,
        },
    },
    {
        "name": "Search + Popularity - Simple Search with GIN Index (Offset Pagination)",
        "group": "Search + Popularity",
        "size": dataset_len,
        "query": {
            "search_query_text": "speaker",
            "search_type": "simple",
            "sort_by_column": "bought_in_last_month",
            "sort_order": "desc",
            "page_size": 50,
            "page_number": 1,
        },
    },
    ## Search + Popularity - Full-Text Search (vector search)
    {
        "name": "Search + Popularity - Full-Text Search (vector search) (Keyset Pagination)",
        "group": "Search + Popularity",
        "size": dataset_len,
        "query": {
            "search_query_text": "speaker",
            "search_type": "fts",
            "sort_by_column": "bought_in_last_month",
            "sort_order": "desc",
            "page_size": 50,
            "size": dataset_len,
        },
    },
    {
        "name": "Search + Popularity - Full-Text Search (vector search) (Offset Pagination)",
        "group": "Search + Popularity",
        "size": dataset_len,
        "query": {
            "search_query_text": "speaker",
            "search_type": "fts",
            "sort_by_column": "bought_in_last_month",
            "sort_order": "desc",
            "page_size": 50,
            "page_number": 1,
            "size": dataset_len,
        },
    },
]


def safe_slug(text: str) -> str:
    normalized = re.sub(r"[^a-z0-9]+", "_", text.lower()).strip("_")
    return normalized or "query"


if __name__ == "__main__":
    # load_csv_and_insert_bulk("external-data/processed_amazon_uk_products.csv", limit=dataset_size)
    # # sleep for a bit to ensure all data is inserted before we start fetching
    # print("Data insertion complete. Starting to fetch products and measure response times...")
    # time.sleep(60)
    for query in queries:
        name = query["name"]
        query_params = query["query"]
        response_times[name] = []
        start = time.time()
        page_no = 1

        last_id = -1
        sort_by_column = query_params.get("sort_by_column", "")
        sort_order = query_params.get("sort_order", "asc")
        sort_last_value = None
        page_no = 1
        page_size = query_params.get("page_size", 50)
        count = 1
        limit_offset_pagination = "page_number" in query_params
        next_page = 1 if not limit_offset_pagination else dataset_size // 1000
        page_snapshot_interval = dataset_size // 1000
        if "Search" in name and limit_offset_pagination:
            next_page = 10  # skip only 10 pages for search queries as there may be fewer results
        if "Search" in name:
            page_snapshot_interval = 10  # capture every 10 pages for search queries due to likely smaller result set
        # Get Count of total results for the query to determine how many pages to fetch
        count_response = requests.get(COUNT_API, params=query_params)
        if count_response.status_code != 200:
            print(f"Failed to fetch count for query {name}: {count_response.text}")
            continue
        count_response = count_response.json()
        total_count = count_response.get("count", 0)
        total_pages = (total_count // page_size) + (1 if total_count % page_size > 0 else 0)

        print(f"Total results for query '{name}': {total_count}, Total pages: {total_pages}")

        while count > 0 and page_no <= total_pages:
            query_next = copy.deepcopy(query_params)

            if limit_offset_pagination:
                query_next.update(
                    {
                        "sort_by_column": sort_by_column,
                        "sort_order": sort_order,
                        "page_number": page_no,
                    }
                )
            else:
                query_next.update(
                    {
                        "last_id": last_id,
                        "sort_by_column": sort_by_column,
                        "sort_order": sort_order,
                        "sort_last_value": sort_last_value,
                    }
                )
            start = time.time()
            # print(query_next)
            response = requests.get(FETCH_API, params=query_next)
            if response.status_code != 200:
                print(f"Failed to fetch products for query {name}: {response.text}")
                break
            end = time.time()
            print(f"Query: {name} Page: {page_no}: {end - start:.4f} seconds")
            if limit_offset_pagination:
                response_times[name].append({"page": page_no, "response_time_ms": round((end - start) * 1000)})
            elif not limit_offset_pagination and page_no == 1 or page_no % page_snapshot_interval == 1 or page_no == total_pages:
                response_times[name].append({"page": page_no, "response_time_ms": round((end - start) * 1000)})
            if limit_offset_pagination and page_no >= total_pages:
                break  # Stop if we've fetched all pages for offset pagination
            page_no += next_page
            if limit_offset_pagination and page_no > total_pages:
                page_no = total_pages  # Ensure that we go to the last page if we overshoot

            response = response.json()

            products = response.get("products", [])
            product_ids = [p["id"] for p in products]
            # print(f"Fetched {len(product_ids)} products: {product_ids}")

            last_id = response.get("last_id", None)
            sort_by_column = response.get("sort_by_column", "")
            sort_order = response.get("sort_order", "asc")
            sort_last_value = response.get("sort_last_value", None)
            count = response.get("count", 0)

    with open(PATH, "w") as f:
        json.dump(response_times, f, indent=4)
