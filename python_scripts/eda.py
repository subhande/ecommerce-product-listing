import json
import os
import random

import pandas as pd
import requests
from tqdm import tqdm

PATH = "external-data/amz_uk_processed_data.csv"
# PATH = "external-data/Amazon Products Dataset 2023 (1.4M Products)/amazon_products.csv"

def load_data(path):
    """Load the dataset from the specified CSV file."""
    return pd.read_csv(path)




def get_brand_name(product_details: str) -> str:
    payload = {
        # "model": "google/gemma-3-4b",
        "model": "google/gemma-3-1b",
        "messages": [
            {
                "role": "system",
                "content": """Extract the brand name from the given product details. Respond with JSON like {\"brand_name\": \"...\"}.
                Sometimes brand name might not be present in the details, in that case check your prior knowledge to infer the brand name.
                """
            },
            {
                "role": "user",
                "content": product_details
            }
        ],
        "temperature": 0.7,
        "max_tokens": 100
    }

    response = requests.post(
        "http://localhost:1234/v1/chat/completions",
        headers={
            # "Authorization": f"Bearer {os.environ['LM_API_TOKEN']}",
            "Content-Type": "application/json"
        },
        json=payload
    )
    response.raise_for_status()
    data = response.json()
    content = data["choices"][0]["message"]["content"].strip()

    try:
        content = content.replace("```json", "").replace("```", "").replace("\n", "")
        parsed = json.loads(content)
        brand_name = parsed.get("brand_name", "").strip()
    except json.JSONDecodeError:
        brand_name = content

    # print(json.dumps(data, indent=2))
    return brand_name


country_to_currency_map = {
    "UK": "GBP",
    "US": "USD",
    "CA": "CAD",
    "IN": "INR",
}


def basic_eda(path: str , country: str ) -> None:
    """Perform basic exploratory data analysis on the DataFrame."""

    df = pd.read_csv(path)

    currency = country_to_currency_map.get(country.upper(), "USD")

    print("First 5 rows of the dataset:")
    print(df.head())

    print("\nDataset Info:")
    print(df.info())

    print("\nStatistical Summary:")
    print(df.describe())

    print("\nMissing Values in Each Column:")
    print(df.isnull().sum())

    print("\nData Types of Each Column:")
    print(df.dtypes)

    data = []

    for idx, row in tqdm(df.iterrows(), total=len(df)):
        # print(f"\nRow {idx} details:")
        # pr
        # product_detils = ""
        # for col in df.columns:
        #     # print(f"{col}: {row[col]}")
        #     product_detils += f"{col}: {row[col]}\n"
        # brand_name = get_brand_name(product_detils)
        # row["brand_name"] = brand_name
        # print(f"Extracted Brand Name: {brand_name} ||| {row['title']}")

        row = {
            "title": row["title"] if not pd.isna(row["title"]) else "",
            "asin": row["asin"] if not pd.isna(row["asin"]) else "",
            "description": "",
            "category": row["categoryName"] if not pd.isna(row["categoryName"]) else "",
            "brand": "",
            "image_url": row["imgUrl"] if not pd.isna(row["imgUrl"]) else "",
            "product_url": row["productURL"] if not pd.isna(row["productURL"]) else "",
            "price": float(row["price"]) if not pd.isna(row["price"]) else 0.0,
            "currency": currency,
            "country": country,
            "stock": random.randint(0, 100),
            "avg_rating": float(row["stars"]) if not pd.isna(row["stars"]) else 0.0,
            "review_count": int(row["reviews"]) if not pd.isna(row["reviews"]) else 0,
            "bought_in_last_month": int(row["boughtInLastMonth"]) if not pd.isna(row["boughtInLastMonth"]) else 0,
            "is_best_seller": bool(row["isBestSeller"]) if not pd.isna(row["isBestSeller"]) else False,
        }

        data.append(row)

    df_new = pd.DataFrame(data)
    # description column is missing, fill with empty strings and set as string dtype
    df_new["description"] = df_new["description"].astype(str)
    df_new["brand"] = df_new["brand"].astype(str)
    output_path = f"external-data/processed_amazon_{country.lower()}_products.csv"
    df_new.to_csv(output_path, index=False)
    print(f"\nProcessed data saved to {output_path}")

    # productURL = df["productURL"].tolist()

    # with open("external-data/sample_product_urls.json", "w") as f:

    #     json.dump(productURL, f, indent=4)

if __name__ == "__main__":
    # data = load_data()
    # basic_eda(data)
    # product_details = """
    # asin: B09B96TG33
    # title: Echo Dot (5th generation, 2022 release) | Big vibrant sound Wi-Fi and Bluetooth smart speaker with Alexa | Charcoal
    # imgUrl: https://m.media-amazon.com/images/I/71C3lbbeLsL._AC_UL320_.jpg
    # productURL: https://www.amazon.co.uk/dp/B09B96TG33
    # stars: 4.7
    # reviews: 15308
    # price: 21.99
    # isBestSeller: False
    # boughtInLastMonth: 0
    # categoryName: Hi-Fi Speakers
    # """
    # get_brand_name(product_details)
    UK_PRODUCTS_PATH = "external-data/amz_uk_processed_data.csv"
    basic_eda(UK_PRODUCTS_PATH, country="UK")
