

# Designing an E-commerce Product Listing System



# Product Listing Queries
- Search Products by Name, Category, Price Range
  - e.g. "Smartphones under €500", "smartphones from Samsung", "laptops under €1000"
- Filter by Price, Rating, Brand, Reviews More than X
- Sort By Price (Low to High, High to Low), Rating, Popularity



## Day 1 Architecture (upto 1k products)

## Day 100 Architecture (upto 10k - 100k products)

## Final Architecture (upto 1M+ products)


### PG_TEXTSEARCH Alternate to ElasticSearch
https://youtu.be/XEiQV4zRC-U


# Benchmark:
- Sort
- Filter + Sort
- Search + Sort
- Search + Filter + Sort


## Indexes

### Global Sorting & Keyset Pagination
These indexes support the "All Products" view or search results where no specific category/brand filter is applied. They match your ORDER BY clause exactly.

- **Popularity** : `CREATE INDEX idx_products_popular_keyset ON products (bought_in_last_month DESC, id DESC) WHERE stock > 0;`

- **Price (Low to High)**: `CREATE INDEX idx_products_price_asc_keyset ON products (price ASC, id ASC) WHERE stock > 0;`

- **Rating**: `CREATE INDEX idx_products_rating_keyset ON products (avg_rating DESC, id DESC) WHERE stock > 0;`

- **Newest/Updated**: `CREATE INDEX idx_products_updated_keyset ON products (updated_at DESC, id DESC) WHERE stock > 0;`



## Category-Specific Filters
Users frequently browse by category. These indexes allow PostgreSQL to jump to a specific category and immediately have the items in the requested sort order.

- **Category + Popularity**: `CREATE INDEX idx_products_cat_popular ON products (category, bought_in_last_month DESC, id DESC) WHERE stock > 0;`

- **Category + Price**: `CREATE INDEX idx_products_cat_price ON products (category, price ASC, id ASC) WHERE stock > 0;`

- **Category + Rating**: `CREATE INDEX idx_products_cat_rating ON products (category, avg_rating DESC, id DESC) WHERE stock > 0;`

- **Category + Newest**: `CREATE INDEX idx_products_cat_updated ON products (category, updated_at DESC, id DESC) WHERE stock > 0;`


## Brand-Specific Filters

Similar to categories, brand-based filtering is a primary use case in your ProductFilter.

- **Brand + Popularity**: `CREATE INDEX idx_products_brand_popular ON products (brand, bought_in_last_month DESC, id DESC) WHERE stock > 0;`

- **Brand + Price**: `CREATE INDEX idx_products_brand_price ON products (brand, price ASC, id ASC) WHERE stock > 0;`

- **Brand + Rating**: `CREATE INDEX idx_products_brand_rating ON products (brand, avg_rating DESC, id DESC) WHERE stock > 0;`

- **Brand + Newest**: `CREATE INDEX idx_products_brand_updated ON products (brand, updated_at DESC, id DESC) WHERE stock > 0;`


## Full-Text Search Optimization

Your repository uses ILIKE with wildcards (e.g., %query%), which makes standard B-Tree indexes useless. You should enable the pg_trgm extension and use a GIN index.

- **Enable Extension**: `CREATE EXTENSION IF NOT EXISTS pg_trgm;`

- **Search Index**: `CREATE INDEX idx_products_search_trgm ON products USING gin (title gin_trgm_ops, description gin_trgm_ops);`



# Result

### Sort

| Dataset Size                                                         | 1k Products | 10k Products | 100k Products | 1M+ Products |
| -------------------------------------------------------------------- | ----------- | ------------ | ------------- | ------------ |
| Popularity [Without Index]                                           | 10ms        | 50ms         | 200ms         | 1s           |
| Popularity [With Index]                                              | 1ms         | 2ms          | 5ms           | 20ms         |
| Popularity [With Index Last Page] (Limit Offset Pagination)          | 1ms         | 2ms          | 5ms           | 20ms         |
| Polularity [With Index First Page] (Keyset Pagination)               | 1ms         | 2ms          | 5ms           | 20ms         |
| Price (Low to High) [Without Index]                                  | 15ms        | 100ms        | 500ms         | 5s           |
| Price (Low to High) [With Index]                                     | 1ms         |              | 3ms           | 15ms         |
| Price (Low to High) [With Index Last Page] (Limit Offset Pagination) | 1ms         | 3ms          | 10ms          | 100ms        |
| Price (Low to High) [With Index First Page] (Keyset Pagination)      | 1ms         | 3ms          | 10ms          | 15ms         |
| Rating [Without Index]                                               |             | 20ms         | 150ms         | 700ms        | 10s  |
| Rating [With Index]                                                  | 1ms         | 3ms          |               | 10ms         | 50ms |
| Rating [With Index Last Page] (Limit Offset Pagination)              | 1ms         | 3ms          | 10ms          | 100ms        |
| Rating [With Index First Page] (Keyset Pagination)                   | 1ms         | 3ms          | 10ms          | 15ms         |
| Newest/Updated [Without Index]                                       | 20ms        | 150ms        | 700ms         | 10s          |
| Newest/Updated [With Index]                                          | 1ms         | 3ms          | 10ms          | 50ms         |
| Newest/Updated [With Index Last Page] (Limit Offset Pagination)      | 1ms         | 3ms          | 10ms          | 100ms        |
| Newest/Updated [With Index First Page] (Keyset Pagination)           | 1ms         | 3ms          | 10ms          | 15ms         |



## Filter

