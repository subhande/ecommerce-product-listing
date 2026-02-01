import json

import requests
from faker import Faker

fake =Faker()

BASE_URL = "http://localhost:8080/api/v1/products"

INSERT_API = f"{BASE_URL}"
INSERT_BULK_API = f"{BASE_URL}/bulk"
FETCH_API = f"{BASE_URL}"

product_categories = [
    "Electronics", "Clothing", "Books", "Home & Kitchen", "Sports", "Toys", "Beauty", "Automotive", "Grocery",
    "Health", "Garden", "Tools", "Office Supplies", "Pet Supplies"
]

def create_product_payload():
    return {
        "name": fake.word().title(),
        "description": fake.text(max_nb_chars=200),
        "price": round(fake.pyfloat(left_digits=3, right_digits=2, positive=True), 2),
        "stock": fake.random_int(min=0, max=1000),
        "category": fake.random.choice(product_categories)
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
        params['category'] = category
    params['min_price'] = min_price
    params['max_price'] = max_price
    response = requests.get(FETCH_API, params=params)
    if response.status_code == 200:
        products = response.json()
        print(f"Fetched {len(products)} Products")
        for product in products:
            print(product)
    else:
        print(f"Failed to fetch products: {response.text}")


if __name__ == "__main__":
    # # Insert a single product
    # single_product_payload = create_product_payload()
    # # print(single_product_payload)
    # insert_product(single_product_payload)

    # Insert bulk products
    bulk_product_payloads = [create_product_payload() for _ in range(10000)]
    insert_bulk_products(bulk_product_payloads)

    # # Fetch products with filters
    # fetch_products(category="Electronics", min_price=50.0, max_price=500.0)
