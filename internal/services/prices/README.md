# Prices Service

This package provides a service for managing product prices loaded from a CSV file.

## Features

- Load prices from CSV files with customizable file paths
- Get price information by product name (case-insensitive)
- Find products by exact price
- Find products within a price range
- List all products and prices

## Usage

### Basic Setup

```go
import "github.com/rogerwesterbo/svennescamping-backend/internal/services/prices"

// Initialize the service with a CSV file path
priceService, err := prices.NewPriceService("/path/to/prices.csv")
if err != nil {
    log.Fatal(err)
}
```

### Environment Configuration

For flexible deployment (local vs Kubernetes), you can use environment variables:

```go
csvPath := "hack/data/prices.csv" // Default local path

// Override with environment variable for Kubernetes
if envPath := os.Getenv("PRICES_CSV_PATH"); envPath != "" {
    csvPath = envPath
}

priceService, err := prices.NewPriceService(csvPath)
```

### Available Methods

#### Get Price by Product

```go
price, err := priceService.GetPriceByProduct("Cabin")
if err != nil {
    // Product not found
}
// price.Product = "Cabin"
// price.Price = 650.0
// price.Currency = "NOK"
```

#### Get Products by Exact Price

```go
products, err := priceService.GetProductsByPrice(410.0)
if err != nil {
    // No products found with this price
}
```

#### Get Products by Price Range

```go
products, err := priceService.GetProductsByPriceRange(400.0, 450.0)
if err != nil {
    // Invalid range or no products found
}
```

#### Get All Products

```go
allProducts := priceService.GetAllProducts()
// Returns []string with all product names
```

#### Get All Prices

```go
allPrices := priceService.GetAllPrices()
// Returns []Price with all price information
```

## CSV Format

The service expects a semicolon-separated CSV file with the following format:

```csv
Product;Price;Currency
Cabin;650;NOK
Bed linen;75;NOK
Caravan/motorhome/tent 1-2 pers;390;NOK
```

### Requirements:

- First row must be headers: `Product;Price;Currency`
- Semicolon (`;`) as delimiter
- Price column must contain valid numeric values
- All rows must have exactly 3 columns

## Deployment

### Local Development

Place your CSV file in `hack/data/prices.csv` or specify a custom path.

### Kubernetes Deployment

Set the `PRICES_CSV_PATH` environment variable to point to your mounted CSV file:

```yaml
env:
  - name: PRICES_CSV_PATH
    value: "/app/data/prices.csv"
```

You can mount the CSV file using ConfigMaps or persistent volumes.

## Testing

Run the tests with:

```bash
go test ./internal/services/prices/... -v
```

## Example

See `cmd/prices-example/main.go` for a complete usage example.

Run the example:

```bash
go run cmd/prices-example/main.go
```
