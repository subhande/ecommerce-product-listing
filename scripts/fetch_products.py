import json
import time, copy

import pandas as pd
import requests
from faker import Faker
from tqdm import tqdm

fake = Faker()

BASE_URL = "http://localhost:8080/api/v1/products"


FETCH_API = f"{BASE_URL}"


queries = [
    ## Global Sorting
    {
        "name": "Popular Products",
        "query": {"sort_by_column": "bought_in_last_month", "sort_order": "desc", "page_size": 20},
    },
    {
        "name": "Price Low to High",
        "query": {"sort_by_column": "price", "sort_order": "asc", "page_size": 20},
    },
    {
        "name": "Rating High to Low",
        "query": {"sort_by_column": "avg_rating", "sort_order": "desc", "page_size": 20},
    },
    {
        "name": "Recently Updated",
        "query": {"sort_by_column": "updated_at", "sort_order": "desc", "page_size": 20},
    },
    ## Category Filtering
    {
        "name": "Filter by Category",
        "query": {
            "category": "Graphics Cards",
            "sort_by_column": "price",
            "sort_order": "asc",
            "page_size": 20,
        },
    },
]

for query in queries:
    name = query["name"]
    query_params = query["query"]
    start = time.time()
    page_no = 1
    response = requests.get(FETCH_API, params=query_params)
    if response.status_code != 200:
        print(f"Failed to fetch products for query {name}: {response.text}")
        continue
    end = time.time()
    print(f"Query: {name} Page: {page_no}: {end - start:.4f} seconds")
    page_no += 1

    response = response.json()

    products = response.get("products", [])
    product_ids = [product["id"] for product in products]
    print(f"Fetched {len(product_ids)} products: {product_ids}")

    last_id = response.get("last_id", None)
    sort_by_column = response.get("sort_by_column", "")
    sort_order = response.get("sort_order", "asc")
    sort_last_value = response.get("sort_last_value", None)
    count = response.get("count", 0)

    while count > 0:
        query_next = copy.deepcopy(query_params)
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
        page_no += 1

        response = response.json()

        products = response.get("products", [])
        product_ids = [p["id"] for p in products]
        print(f"Fetched {len(product_ids)} products: {product_ids}")

        last_id = response.get("last_id", None)
        sort_by_column = response.get("sort_by_column", "")
        sort_order = response.get("sort_order", "asc")
        sort_last_value = response.get("sort_last_value", None)
        count = response.get("count", 0)
