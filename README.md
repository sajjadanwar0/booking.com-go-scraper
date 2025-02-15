# Booking.com Hotel Scraper

A high-performance web scraper written in Go that extracts hotel information from Booking.com. This tool uses ChromeDP for browser automation and supports concurrent scraping with customizable parameters.

## Features

- Scrape hotel details including name, location, and price
- Customizable number of hotels to scrape
- Progress tracking with colored output
- CSV export of results
- Built-in rate limiting and polite delays
- Error handling and retry logic
- Clean and maintainable codebase

## Prerequisites

Before running the scraper, make sure you have the following installed:

- Go 1.16 or higher
- Chrome/Chromium browser

## Installation

1. Clone the repository:
```bash
git clone https://github.com/sajjadanwar0/booking.com-go-scraper.git
cd booking.com-go-scraper
```

2. Install the required dependencies:
```bash
go mod init booking.com-go-scraper
go get github.com/PuerkitoBio/goquery
go get github.com/chromedp/chromedp
go get github.com/fatih/color
```

## Usage

Basic usage with default settings (200 hotels):
```bash
go run main.go "United States"
```

Scrape a specific number of hotels:
```bash
go run main.go -n 300 "France"
```

### Command Line Arguments

- `-n`: Number of hotels to scrape (default: 200)
- Country name: Required positional argument (use quotes for names with spaces)

### Examples

```bash
# Scrape 150 hotels in Spain
go run main.go -n 150 "Spain"

# Scrape 500 hotels in United Kingdom
go run main.go -n 500 "United Kingdom"

# Use default settings for Italy
go run main.go "Italy"
```

## Output

The scraper creates a CSV file named after the country (e.g., `united_states_hotels.csv`) containing:
- Hotel Name
- Location
- Price

Example output format:
```csv
Name,Location,Price
Hilton Garden Inn Times Square,"123 W 42nd St, New York",199
The Plaza Hotel,"5th Avenue, New York",599
...
```

## Features in Detail

### Browser Automation
- Uses ChromeDP for headless browser control
- Handles dynamic content loading
- Manages browser resources efficiently

### Rate Limiting
- Implements polite delays between requests
- Respects website's rate limits
- Prevents IP blocking

### Error Handling
- Graceful handling of network errors
- Retry logic for failed requests
- Clear error messages and logging

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
