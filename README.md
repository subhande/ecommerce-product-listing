

# Designing an E-commerce Product Listing System



# Product Listing Queries
- Search Products by Name, Category, Price Range
  - e.g. "Smartphones under €500", "smartphones from Samsung", "laptops under €1000"
- Filter by Price, Rating, Brand, Reviews More than X
- Sort By Price (Low to High, High to Low), Rating, Popularity

# TODO:
- [x] Day 1 architecture - API + Single PostgreSQL instance
- [x] Benchmarking with different dataset sizes (10k, 100k, 1M products)
- [x] Compare Keyset Pagination vs Offset Pagination
- [ ] Load Test on 10k, 100k, 1M products
- [ ] Read Replicas for scaling reads
- [ ] Caching with Redis for popular queries
- [ ] ElasticSearch for full-text search and complex filtering


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
| Query                                              | Type   | Dataset Size | Total Pages | First Page | Last Page | Avg Response Time (ms) | Median Response Time (ms) | P90 Response Time (ms) | P95 Response Time (ms) | P99 Response Time (ms) |
| -------------------------------------------------- | ------ | ------------ | ----------- | ---------- | --------- | ---------------------- | ------------------------- | ---------------------- | ---------------------- | ---------------------- |
| Popular Products                                   | Keyset | ~10k         | 199         | 4          | 1         | 2.24                   | 2                         | 2                      | 2                      | 2                      |
| Popular Products                                   | Keyset | ~100k        | 1981        | 5          | 3         | 3.24                   | 2                         | 2                      | 3                      | 3                      |
| Popular Products                                   | Keyset | ~1M          | 19804       | 13         | 2         | 4.43                   | 5                         | 3                      | 4                      | 2                      |
| Popular Products                                   | Offset | ~10k         | 199         | 3          | 4         | 4.95                   | 4                         | 5                      | 4                      | 4                      |
| Popular Products                                   | Offset | ~100k        | 1981        | 3          | 22        | 15.14                  | 16                        | 20                     | 21                     | 22                     |
| Popular Products                                   | Offset | ~1M          | 19804       | 10         | 668       | 679.43                 | 1019                      | 625                    | 697                    | 668                    |
| Price Low to High                                  | Keyset | ~10k         | 199         | 3          | 1         | 2.60                   | 2                         | 2                      | 2                      | 2                      |
| Price Low to High                                  | Keyset | ~100k        | 1981        | 2          | 2         | 2.52                   | 2                         | 2                      | 2                      | 2                      |
| Price Low to High                                  | Keyset | ~1M          | 19804       | 8          | 3         | 3.67                   | 3                         | 4                      | 5                      | 3                      |
| Price Low to High                                  | Offset | ~10k         | 199         | 5          | 6         | 7.05                   | 4                         | 6                      | 6                      | 6                      |
| Price Low to High                                  | Offset | ~100k        | 1981        | 3          | 52        | 27.71                  | 27                        | 44                     | 49                     | 52                     |
| Price Low to High                                  | Offset | ~1M          | 19804       | 14         | 3501      | 1669.19                | 1582                      | 3101                   | 4230                   | 3501                   |
| Rating High to Low                                 | Keyset | ~10k         | 199         | 2          | 2         | 2.13                   | 2                         | 2                      | 2                      | 2                      |
| Rating High to Low                                 | Keyset | ~100k        | 1981        | 3          | 2         | 2.57                   | 3                         | 3                      | 3                      | 2                      |
| Rating High to Low                                 | Keyset | ~1M          | 19804       | 6          | 3         | 3.52                   | 3                         | 2                      | 3                      | 3                      |
| Rating High to Low                                 | Offset | ~10k         | 199         | 3          | 6         | 6.29                   | 4                         | 6                      | 6                      | 6                      |
| Rating High to Low                                 | Offset | ~100k        | 1981        | 3          | 37        | 28.00                  | 23                        | 37                     | 36                     | 37                     |
| Rating High to Low                                 | Offset | ~1M          | 19804       | 6          | 1737      | 1128.90                | 1979                      | 1801                   | 1829                   | 1737                   |
| Recently Updated                                   | Keyset | ~10k         | 199         | 2          | 2         | 2.33                   | 2                         | 2                      | 2                      | 2                      |
| Recently Updated                                   | Keyset | ~100k        | 1981        | 2          | 3         | 2.57                   | 3                         | 3                      | 2                      | 3                      |
| Recently Updated                                   | Keyset | ~1M          | 19804       | 9          | 3         | 3.86                   | 3                         | 3                      | 3                      | 3                      |
| Recently Updated                                   | Offset | ~10k         | 199         | 3          | 5         | 5.14                   | 3                         | 4                      | 4                      | 5                      |
| Recently Updated                                   | Offset | ~100k        | 1981        | 3          | 19        | 12.00                  | 12                        | 17                     | 17                     | 19                     |
| Recently Updated                                   | Offset | ~1M          | 19804       | 13         | 421       | 213.14                 | 198                       | 361                    | 361                    | 421                    |
| Filter by Category - Price Low to High             | Keyset | ~10k         | 1           | 2          | 2         | 2.00                   | 2                         | 2                      | 2                      | 2                      |
| Filter by Category - Price Low to High             | Keyset | ~100k        | 2           | 2          | 2         | 2.00                   | 2                         | 2                      | 2                      | 2                      |
| Filter by Category - Price Low to High             | Keyset | ~1M          | 4           | 4          | 2         | 3.00                   | 2                         | 2                      | 2                      | 2                      |
| Filter by Category - Price Low to High             | Offset | ~10k         | 1           | 2          | 2         | 2.00                   | 2                         | 2                      | 2                      | 2                      |
| Filter by Category - Price Low to High             | Offset | ~100k        | 2           | 2          | 2         | 2.00                   | 2                         | 2                      | 2                      | 2                      |
| Filter by Category - Price Low to High             | Offset | ~1M          | 4           | 3          | 2         | 2.50                   | 2                         | 2                      | 2                      | 2                      |
| Filter by Category - Popularity                    | Keyset | ~10k         | 1           | 2          | 2         | 2.00                   | 2                         | 2                      | 2                      | 2                      |
| Filter by Category - Popularity                    | Keyset | ~100k        | 2           | 2          | 8         | 5.00                   | 8                         | 8                      | 8                      | 8                      |
| Filter by Category - Popularity                    | Keyset | ~1M          | 4           | 3          | 71        | 37.00                  | 71                        | 71                     | 71                     | 71                     |
| Filter by Category - Popularity                    | Offset | ~10k         | 1           | 2          | 2         | 2.00                   | 2                         | 2                      | 2                      | 2                      |
| Filter by Category - Popularity                    | Offset | ~100k        | 2           | 2          | 2         | 2.00                   | 2                         | 2                      | 2                      | 2                      |
| Filter by Category - Popularity                    | Offset | ~1M          | 4           | 4          | 4         | 4.00                   | 4                         | 4                      | 4                      | 4                      |
| Filter by Category - Newest                        | Keyset | ~10k         | 1           | 2          | 2         | 2.00                   | 2                         | 2                      | 2                      | 2                      |
| Filter by Category - Newest                        | Keyset | ~100k        | 2           | 2          | 2         | 2.00                   | 2                         | 2                      | 2                      | 2                      |
| Filter by Category - Newest                        | Keyset | ~1M          | 4           | 3          | 2         | 2.50                   | 2                         | 2                      | 2                      | 2                      |
| Filter by Category - Newest                        | Offset | ~10k         | 1           | 2          | 2         | 2.00                   | 2                         | 2                      | 2                      | 2                      |
| Filter by Category - Newest                        | Offset | ~100k        | 2           | 2          | 2         | 2.00                   | 2                         | 2                      | 2                      | 2                      |
| Filter by Category - Newest                        | Offset | ~1M          | 4           | 2          | 2         | 2.00                   | 2                         | 2                      | 2                      | 2                      |
| Search + Popularity - Simple Search with GIN Index | Keyset | ~10k         | 4           | 23         | 5         | 13.50                  | 10                        | 5                      | 5                      | 5                      |
| Search + Popularity - Simple Search with GIN Index | Keyset | ~100k        | 42          | 10         | 6         | 6.83                   | 6                         | 6                      | 6                      | 6                      |
| Search + Popularity - Simple Search with GIN Index | Keyset | ~1M          | 441         | 229        | 5         | 12.84                  | 6                         | 8                      | 9                      | 5                      |
| Search + Popularity - Simple Search with GIN Index | Offset | ~10k         | 4           | 19         | 19        | 19.00                  | 19                        | 19                     | 19                     | 19                     |
| Search + Popularity - Simple Search with GIN Index | Offset | ~100k        | 42          | 8          | 8         | 8.00                   | 8                         | 8                      | 8                      | 8                      |
| Search + Popularity - Simple Search with GIN Index | Offset | ~1M          | 441         | 207        | 83        | 108.38                 | 96                        | 89                     | 84                     | 83                     |
| Search + Popularity - Full-Text Search             | vector | ~10k         | 4           | 5          | 3         | 3.75                   | 3                         | 3                      | 3                      | 3                      |
| Search + Popularity - Full-Text Search             | vector | ~100k        | 40          | 9          | 3         | 3.80                   | 3                         | 3                      | 3                      | 3                      |
| Search + Popularity - Full-Text Search             | vector | ~1M          | 413         | 158        | 6         | 9.12                   | 10                        | 6                      | 3                      | 6                      |
| Search + Popularity - Full-Text Search             | vector | ~10k         | 4           | 4          | 4         | 4.00                   | 4                         | 4                      | 4                      | 4                      |
| Search + Popularity - Full-Text Search             | vector | ~100k        | 40          | 8          | 4         | 5.00                   | 4                         | 4                      | 4                      | 4                      |
| Search + Popularity - Full-Text Search             | vector | ~1M          | 413         | 134        | 84        | 100.79                 | 76                        | 89                     | 76                     | 84                     |

## Filter

