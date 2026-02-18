import json

import pandas as pd
import requests
from faker import Faker
from tqdm import tqdm

fake = Faker()

BASE_URL = "http://localhost:8080/api/v1/products"

INSERT_API = f"{BASE_URL}"
INSERT_BULK_API = f"{BASE_URL}/bulk"
FETCH_API = f"{BASE_URL}"

product_categories = [
    "Electronics",
    "Clothing",
    "Books",
    "Home & Kitchen",
    "Sports",
    "Toys",
    "Beauty",
    "Automotive",
    "Grocery",
    "Health",
    "Garden",
    "Tools",
    "Office Supplies",
    "Pet Supplies",
]


def create_product_payload():
    return {
        "name": fake.word().title(),
        "description": fake.text(max_nb_chars=200),
        "price": round(fake.pyfloat(left_digits=3, right_digits=2, positive=True), 2),
        "stock": fake.random_int(min=0, max=1000),
        "category": fake.random.choice(product_categories),
    }


def insert_product(payload):
    response = requests.post(INSERT_API, json=payload)
    if response.status_code == 201:
        print(f"Inserted Product: {response.json()}")
    else:
        print(f"Failed to insert product: {response.text}")


def insert_bulk_products(payloads):
    response = requests.post(INSERT_BULK_API, json=payloads)
    if response.status_code == 201:
        print(f"Inserted Bulk Products: {response.json()}")
    else:
        print(f"Failed to insert bulk products: {response.text}")


def fetch_products(category: str = "", min_price: float = 0.0, max_price: float = float("inf")):
    params = {}
    if category:
        params["category"] = category
    params["min_price"] = min_price
    params["max_price"] = max_price
    response = requests.get(FETCH_API, params=params)
    if response.status_code == 200:
        products = response.json()
        print(f"Fetched {len(products)} Products")
        for product in products:
            print(product)
    else:
        print(f"Failed to fetch products: {response.text}")


def load_csv_and_insert_bulk(path: str, limit: int = 100000):

    df = pd.read_csv(path)
    if limit != -1:
        for category in df["category"].unique():
            category_df = df[df["category"] == category]
            if len(category_df) > (limit // 100):
                category_sample = category_df.sample(n=limit // 100, random_state=42)
            else:
                category_sample = category_df
            if "sampled_df" in locals():
                sampled_df = pd.concat([sampled_df, category_sample], ignore_index=True)
            else:
                sampled_df = category_sample
        df = sampled_df.reset_index(drop=True)

        print(f"Total products after sampling: {len(df)}")

        df = df.sample(n=(limit), random_state=42).reset_index(drop=True)
    print(f"Loaded CSV with {len(df)} products for insertion")
    # Loop through columns and fill based on dtype
    for col in df.columns:
        if str(col) in ["description", "brand"]:
            df[col] = df[col].fillna("")
        else:
            if pd.api.types.is_numeric_dtype(df[col]):
                df[col] = df[col].fillna(0)

            elif pd.api.types.is_bool_dtype(df[col]):
                df[col] = df[col].fillna(False)

            elif pd.api.types.is_datetime64_any_dtype(df[col]):
                df[col] = df[col].fillna(method="ffill")  # or leave as is

            else:  # object / string
                df[col] = df[col].fillna("")
    payloads = []
    for idx, row in tqdm(df.iterrows(), total=len(df)):
        row = dict(row)
        try:
            row = json.loads(json.dumps(row, default=str))
            payloads.append(row)
        except Exception as e:
            print(f"Skipping row {idx} due to serialization error: {e}")
            # print(row)

        if len(payloads) == 1000:
            # print(payloads)
            insert_bulk_products(payloads)
            payloads.clear()
            print(f"Inserted {idx + 1} products so far...")
    if payloads:
        insert_bulk_products(payloads)


if __name__ == "__main__":
    # # Insert a single product
    # single_product_payload = create_product_payload()
    # single_product_payload = {'title': 'Peowuieu Replacement Head Optical Pick-Up Lens with Bracket for -210A Player', 'asin': 'B0CHVQ63P3', 'description': "", 'category': 'CD, Disc & Tape Players', 'brand': "", 'image_url': 'https://m.media-amazon.com/images/I/51+ygjPR+uS._AC_UL320_.jpg', 'product_url': 'https://www.amazon.co.uk/dp/B0CHVQ63P3', 'price': 14.53, 'currency': 'GBP', 'country': 'UK', 'stock': 0, 'avg_rating': 0.0, 'review_count': 0, 'bought_in_last_month': 0, 'is_best_seller': False}
    # print(single_product_payload)
    # insert_product(single_product_payload)

    # # Insert bulk products
    # bulk_product_payloads = [create_product_payload() for _ in range(10000)]
    # insert_bulk_products(bulk_product_payloads)

    # # Fetch products with filters
    # fetch_products(category="Electronics", min_price=50.0, max_price=500.0)

    load_csv_and_insert_bulk("external-data/processed_amazon_uk_products.csv", limit=1000000)
