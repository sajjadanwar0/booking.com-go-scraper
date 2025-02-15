package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/chromedp"
	"github.com/fatih/color"
	"golang.org/x/net/context"
	"log"
	"os"
	"strings"
	"sync"
	"time"
)

type Hotel struct {
	Name     string
	Location string
	Price    string
}

type Scraper struct {
	baseURL   string
	hotels    []Hotel
	mutex     sync.Mutex
	maxHotels int
}

func NewScraper(country string, maxHotels int) *Scraper {
	countryFormatted := strings.ReplaceAll(country, " ", "+")
	baseURL := fmt.Sprintf("https://www.booking.com/searchresults.html?ss=%s&dest_type=country&nflt=&order=popularity", countryFormatted)

	return &Scraper{
		baseURL:   baseURL,
		hotels:    make([]Hotel, 0),
		maxHotels: maxHotels,
	}
}

func (s *Scraper) scrapePage(ctx context.Context, pageNum int) ([]Hotel, error) {
	var hotels []Hotel

	url := fmt.Sprintf("%s&offset=%d", s.baseURL, pageNum*25)
	color.Blue("Accessing URL: %s", url)

	ctx, cancel := chromedp.NewContext(ctx)
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 45*time.Second)
	defer cancel()

	var htmlContent string
	err := chromedp.Run(ctx,
		chromedp.EmulateViewport(1920, 1080),
		chromedp.Navigate(url),
		chromedp.WaitVisible(`div[data-testid="property-card"]`, chromedp.ByQuery),
		chromedp.Evaluate(`window.scrollTo(0, document.body.scrollHeight)`, nil),
		chromedp.Sleep(2*time.Second),
		chromedp.OuterHTML("html", &htmlContent),
	)

	if err != nil {
		return nil, fmt.Errorf("navigation error: %v", err)
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		return nil, fmt.Errorf("parsing error: %v", err)
	}

	doc.Find("div[data-testid='property-card']").Each(func(i int, s *goquery.Selection) {
		// Don't add more hotels if we've already found 25 (page limit)
		if len(hotels) >= 25 {
			return
		}

		hotel := Hotel{}

		// Extract hotel name
		if name := s.Find("div[data-testid='title']").Text(); name != "" {
			hotel.Name = strings.TrimSpace(name)
		} else if name := s.Find("div.a23c043802").Text(); name != "" {
			hotel.Name = strings.TrimSpace(name)
		}

		// Extract location
		if location := s.Find("span[data-testid='address']").Text(); location != "" {
			hotel.Location = strings.TrimSpace(location)
		} else if location := s.Find("span.f4bd0794db").Text(); location != "" {
			hotel.Location = strings.TrimSpace(location)
		}

		// Extract price
		if price := s.Find("span[data-testid='price-and-discounted-price']").Text(); price != "" {
			price = strings.TrimSpace(strings.ReplaceAll(price, "US$", ""))
			hotel.Price = strings.ReplaceAll(price, ",", "")
		} else if price := s.Find("span.fcab3ed991.fbd1d3018c").Text(); price != "" {
			price = strings.TrimSpace(strings.ReplaceAll(price, "US$", ""))
			hotel.Price = strings.ReplaceAll(price, ",", "")
		}

		if hotel.Name != "" {
			hotels = append(hotels, hotel)
			color.Green("Found hotel: %s", hotel.Name)
		}
	})

	if len(hotels) > 0 {
		color.Yellow("Found %d hotels on page %d", len(hotels), pageNum+1)
	} else {
		color.Red("No hotels found on page %d", pageNum+1)
	}

	return hotels, nil
}

func (s *Scraper) saveToCSV(filename string) error {
	if len(s.hotels) == 0 {
		return fmt.Errorf("no hotels found to save")
	}

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			println(err)
		}
	}(file)

	writer := csv.NewWriter(file)
	defer writer.Flush()

	if err := writer.Write([]string{"Name", "Location", "Price"}); err != nil {
		return err
	}

	for _, hotel := range s.hotels {
		if err := writer.Write([]string{hotel.Name, hotel.Location, hotel.Price}); err != nil {
			return err
		}
	}

	return nil
}

func (s *Scraper) Start() error {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("disable-web-security", true),
		chromedp.Flag("disable-notifications", true),
		chromedp.Flag("disable-popup-blocking", true),
		chromedp.Flag("ignore-certificate-errors", true),
		chromedp.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"),
	)

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	ctx, cancel := chromedp.NewContext(allocCtx,
		chromedp.WithLogf(log.Printf),
	)
	defer cancel()

	page := 0
	for len(s.hotels) < s.maxHotels {
		color.Cyan("Scraping page %d...", page+1)

		hotels, err := s.scrapePage(ctx, page)
		if err != nil {
			color.Red("Error scraping page %d: %v", page+1, err)
			continue
		}

		if len(hotels) == 0 {
			color.Yellow("No more hotels found on page %d. Stopping.", page+1)
			break
		}

		s.mutex.Lock()

		remainingSlots := s.maxHotels - len(s.hotels)
		if remainingSlots > len(hotels) {
			s.hotels = append(s.hotels, hotels...)
		} else {
			s.hotels = append(s.hotels, hotels[:remainingSlots]...)
		}
		currentCount := len(s.hotels)
		s.mutex.Unlock()

		color.Green("Total hotels found: %d/%d", currentCount, s.maxHotels)

		if currentCount >= s.maxHotels {
			color.Yellow("Reached target number of hotels. Stopping.")
			break
		}

		page++
		time.Sleep(5 * time.Second)
	}

	if len(s.hotels) == 0 {
		return fmt.Errorf("no hotels were found during scraping")
	}

	return nil
}

func main() {
	// Define command line flags
	maxHotels := flag.Int("n", 200, "Number of hotels to scrape")
	flag.Parse()

	// Get the country from remaining arguments
	args := flag.Args()
	if len(args) < 1 {
		color.Red("Please provide a country name")
		color.Yellow("Usage: go run main.go [-n number_of_hotels] \"country name\"")
		color.Yellow("Example: go run main.go -n 300 \"United States\"")
		os.Exit(1)
	}

	country := args[0]
	outputFile := strings.ToLower(strings.ReplaceAll(country, " ", "_")) + "_hotels.csv"

	color.Cyan("Starting booking.com scraper")
	color.Cyan("Country: %s", country)
	color.Cyan("Target number of hotels: %d", *maxHotels)
	color.Cyan("Output file: %s", outputFile)

	scraper := NewScraper(country, *maxHotels)

	start := time.Now()
	if err := scraper.Start(); err != nil {
		log.Fatal(err)
	}

	if err := scraper.saveToCSV(outputFile); err != nil {
		log.Fatal(err)
	}

	color.Green("✓ Scraping completed in %v", time.Since(start))
	color.Green("✓ Total hotels scraped: %d", len(scraper.hotels))
	color.Green("✓ Results saved to: %s", outputFile)
}
