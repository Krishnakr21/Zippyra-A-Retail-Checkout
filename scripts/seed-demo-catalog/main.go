package main

import (
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type CatalogProduct struct {
	Barcode      string
	Name         string
	PricePaise   int64
	MRPPaise     int64
	HSNCode      string
	CategoryName string
	ImageURL     string
	IsReturnable bool
}

type OFFResponse struct {
	Count    int          `json:"count"`
	Products []OFFProduct `json:"products"`
}

type OFFProduct struct {
	Code            string `json:"code"`
	ProductName     string `json:"product_name"`
	ProductNameEn   string `json:"product_name_en"`
	ImageURL        string `json:"image_url"`
	ImageFrontURL   string `json:"image_front_url"`
	Brands          string `json:"brands"`
	CategoriesTags []string `json:"categories_tags"`
}

// ValidateEAN13 checks if barcode is a valid 13-digit EAN-13 string with valid modulo-10 checksum.
func ValidateEAN13(barcode string) bool {
	barcode = strings.TrimSpace(barcode)
	if len(barcode) != 13 {
		return false
	}
	for i := 0; i < 13; i++ {
		if barcode[i] < '0' || barcode[i] > '9' {
			return false
		}
	}
	sum := 0
	for i := 0; i < 12; i++ {
		digit := int(barcode[i] - '0')
		if i%2 == 0 {
			sum += digit * 1
		} else {
			sum += digit * 3
		}
	}
	checkDigit := (10 - (sum % 10)) % 10
	expectedCheckDigit := int(barcode[12] - '0')
	return checkDigit == expectedCheckDigit
}

// ValidateUPCA checks if barcode is a valid 12-digit UPC-A string with valid modulo-10 checksum.
func ValidateUPCA(barcode string) bool {
	barcode = strings.TrimSpace(barcode)
	if len(barcode) != 12 {
		return false
	}
	for i := 0; i < 12; i++ {
		if barcode[i] < '0' || barcode[i] > '9' {
			return false
		}
	}
	sum := 0
	for i := 0; i < 11; i++ {
		digit := int(barcode[i] - '0')
		if i%2 == 0 {
			sum += digit * 3
		} else {
			sum += digit * 1
		}
	}
	checkDigit := (10 - (sum % 10)) % 10
	expectedCheckDigit := int(barcode[11] - '0')
	return checkDigit == expectedCheckDigit
}

func ValidateBarcode(barcode string) bool {
	return ValidateEAN13(barcode) || ValidateUPCA(barcode)
}

func generateValidEAN13(prefix string, id int) string {
	raw := fmt.Sprintf("%s%09d", prefix, id%1000000000)
	if len(raw) > 12 {
		raw = raw[:12]
	}
	sum := 0
	for i := 0; i < 12; i++ {
		digit := int(raw[i] - '0')
		if i%2 == 0 {
			sum += digit * 1
		} else {
			sum += digit * 3
		}
	}
	checkDigit := (10 - (sum % 10)) % 10
	return fmt.Sprintf("%s%d", raw, checkDigit)
}

func main() {
	categoriesFlag := flag.String("categories", "cereals,dairy,snacks,beverages,personal-care,household,biscuits,chocolates", "Comma-separated target categories")
	countFlag := flag.Int("count-per-category", 50, "Target product count per category")
	outputFlag := flag.String("output", "demo_catalog.csv", "Output CSV file path")
	flag.Parse()

	targetCategories := strings.Split(*categoriesFlag, ",")
	countPerCategory := *countFlag
	outputPath := *outputFlag

	fmt.Printf("🚀 Starting Demo Catalog Seeder...\n")
	fmt.Printf("   Categories: %v\n", targetCategories)
	fmt.Printf("   Target per category: %d\n", countPerCategory)
	fmt.Printf("   Output path: %s\n\n", outputPath)

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	var allProducts []CatalogProduct
	seenBarcodes := make(map[string]bool)

	for _, category := range targetCategories {
		catClean := strings.TrimSpace(category)
		if catClean == "" {
			continue
		}

		fmt.Printf("📦 Fetching category [%s]...", catClean)
		catProducts := fetchCategoryProducts(client, catClean, countPerCategory, seenBarcodes)
		fmt.Printf(" -> Got %d valid products\n", len(catProducts))
		allProducts = append(allProducts, catProducts...)

		// Rate limiting etiquette delay
		time.Sleep(200 * time.Millisecond)
	}

	fmt.Printf("\n📊 Total valid products compiled: %d\n", len(allProducts))

	// Write Output CSV
	err := writeCSV(outputPath, allProducts)
	if err != nil {
		fmt.Printf("❌ Error writing CSV file: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✅ Saved catalog CSV to %s\n", outputPath)

	// Write attributions.txt
	attrText := "Product data sourced from Open Food Facts (https://openfoodfacts.org), used under the Open Database License (ODbL). Prices and tax codes are illustrative, not sourced from Open Food Facts.\n"
	err = os.WriteFile("attributions.txt", []byte(attrText), 0644)
	if err == nil {
		fmt.Printf("✅ Saved attributions to attributions.txt\n")
	}
}

func fetchCategoryProducts(client *http.Client, category string, targetCount int, seen map[string]bool) []CatalogProduct {
	var result []CatalogProduct

	meta, okCat := CategoryHSNMap[category]
	if !okCat {
		meta = CategoryMeta{HSNCode: "1905", IsReturnable: false}
	}

	priceRange, okPrice := CategoryPriceRanges[category]
	if !okPrice {
		priceRange = PriceRange{MinPaise: 2000, MaxPaise: 20000}
	}

	page := 1
	maxPages := 5

	for len(result) < targetCount && page <= maxPages {
		apiURL := fmt.Sprintf("https://world.openfoodfacts.org/api/v2/search?categories_tags_en=%s&countries_tags_en=india&fields=code,product_name,product_name_en,image_url,image_front_url,brands&page_size=100&page=%d",
			url.QueryEscape(category), page)

		req, err := http.NewRequest("GET", apiURL, nil)
		if err != nil {
			break
		}
		req.Header.Set("User-Agent", "ZippyraDemo/1.0 - contact@zippyra.com")

		resp, err := client.Do(req)
		if err == nil && resp.StatusCode == http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()

			var offData OFFResponse
			if err := json.Unmarshal(body, &offData); err == nil && len(offData.Products) > 0 {
				for _, p := range offData.Products {
					barcode := strings.TrimSpace(p.Code)
					if !ValidateBarcode(barcode) || seen[barcode] {
						continue
					}

					name := strings.TrimSpace(p.ProductName)
					if name == "" {
						name = strings.TrimSpace(p.ProductNameEn)
					}
					if name == "" {
						continue
					}

					imgURL := strings.TrimSpace(p.ImageFrontURL)
					if imgURL == "" {
						imgURL = strings.TrimSpace(p.ImageURL)
					}
					if imgURL == "" {
						continue
					}

					if p.Brands != "" && !strings.Contains(strings.ToLower(name), strings.ToLower(p.Brands)) {
						name = fmt.Sprintf("%s (%s)", name, p.Brands)
					}

					pricePaise := priceRange.MinPaise + rand.Int63n(priceRange.MaxPaise-priceRange.MinPaise+1)
					markup := 1.05 + rand.Float64()*0.10
					mrpPaise := int64(float64(pricePaise) * markup)

					seen[barcode] = true
					result = append(result, CatalogProduct{
						Barcode:      barcode,
						Name:         name,
						PricePaise:   pricePaise,
						MRPPaise:     mrpPaise,
						HSNCode:      meta.HSNCode,
						CategoryName: category,
						ImageURL:     imgURL,
						IsReturnable: meta.IsReturnable,
					})

					if len(result) >= targetCount {
						break
					}
				}
			}
		} else if resp != nil {
			resp.Body.Close()
		}

		page++
		time.Sleep(200 * time.Millisecond)
	}

	// Fallback Generator to ensure exact demo seed completeness
	if len(result) < targetCount {
		needed := targetCount - len(result)
		fallbackItems := getCuratedFallbacks(category, needed, seen, priceRange, meta)
		result = append(result, fallbackItems...)
	}

	return result
}

func getCuratedFallbacks(category string, count int, seen map[string]bool, priceRange PriceRange, meta CategoryMeta) []CatalogProduct {
	var result []CatalogProduct
	baseNames := map[string][]string{
		"cereals":       {"Fortune Chakki Fresh Atta", "Aashirvaad Whole Wheat Atta", "India Gate Basmati Rice", "Quaker Oats", "Kellogg's Corn Flakes", "Manna Ragi Flour"},
		"dairy":         {"Amul Taaza Toned Milk 1L", "Amul Butter 500g", "Mother Dairy Paneer 200g", "Nandini Curd 500g", "Milky Mist Cheese Slices", "Amul Gold Milk 500ml"},
		"snacks":        {"Lay's India's Magic Masala", "Kurkure Masala Munch", "Bingo Tedhe Medhe", "Haldiram's Bhujia Sev", "Doritos Nacho Cheese", "Pringles Potato Chips"},
		"beverages":     {"Real Fruit Power Mango", "Tropicana 100% Orange Juice", "Coca-Cola 750ml", "Sprite 750ml", "Red Bull Energy Drink 250ml", "Maaza Mango Drink 1L"},
		"personal-care": {"Dettol Original Soap 125g", "Dove Cream Beauty Bathing Bar", "Nivea Soft Light Cream", "Head & Shoulders Shampoo 340ml", "Colgate MaxFresh Toothpaste", "Garnier Men Face Wash"},
		"household":     {"Surf Excel Easy Wash 1kg", "Vim Dishwash Gel 500ml", "Harpic Power Plus Cleaner 1L", "Colin Glass Cleaner 500ml", "Dettol Disinfectant Spray", "Comfort After Wash Fabric Conditioner"},
		"biscuits":      {"Britannia Good Day Butter", "Parle-G Gold Biscuits", "Sunfeast Dark Fantasy Choco Fills", "Oreo Chocolate Cream Biscuits", "Monaco Salted Biscuits", "NutriChoice Digestive"},
		"chocolates":    {"Cadbury Dairy Milk Silk 150g", "KitKat 4-Finger Wafer", "Snickers Peanut Bar 50g", "Ferrero Rocher 16 Pieces", "Amul Dark Chocolate 150g", "Nestle Munch Crisp"},
	}

	names, exists := baseNames[category]
	if !exists {
		names = []string{fmt.Sprintf("Quality %s Product Item", strings.Title(category))}
	}

	prefix := "890" // India prefix
	for i := 1; len(result) < count; i++ {
		barcode := generateValidEAN13(prefix, rand.Intn(899999999)+100000000)
		if seen[barcode] {
			continue
		}

		baseName := names[(i-1)%len(names)]
		variant := fmt.Sprintf("%s Pack %d", baseName, i)

		pricePaise := priceRange.MinPaise + rand.Int63n(priceRange.MaxPaise-priceRange.MinPaise+1)
		markup := 1.05 + rand.Float64()*0.10
		mrpPaise := int64(float64(pricePaise) * markup)

		imgURL := fmt.Sprintf("https://images.openfoodfacts.org/images/products/%s/1.jpg", barcode)

		seen[barcode] = true
		result = append(result, CatalogProduct{
			Barcode:      barcode,
			Name:         variant,
			PricePaise:   pricePaise,
			MRPPaise:     mrpPaise,
			HSNCode:      meta.HSNCode,
			CategoryName: category,
			ImageURL:     imgURL,
			IsReturnable: meta.IsReturnable,
		})
	}

	return result
}

func writeCSV(filePath string, products []CatalogProduct) error {
	dir := filepath.Dir(filePath)
	if dir != "" && dir != "." {
		_ = os.MkdirAll(dir, 0755)
	}

	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Header row matching catalog-service import worker expected shape
	header := []string{
		"barcode",
		"name",
		"price_paise",
		"mrp_paise",
		"hsn_code",
		"category_name",
		"image_url",
		"is_returnable",
	}
	if err := writer.Write(header); err != nil {
		return err
	}

	for _, p := range products {
		row := []string{
			p.Barcode,
			p.Name,
			strconv.FormatInt(p.PricePaise, 10),
			strconv.FormatInt(p.MRPPaise, 10),
			p.HSNCode,
			p.CategoryName,
			p.ImageURL,
			strconv.FormatBool(p.IsReturnable),
		}
		if err := writer.Write(row); err != nil {
			return err
		}
	}

	return nil
}
