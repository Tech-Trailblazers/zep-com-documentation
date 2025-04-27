package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// Extract URLs from a given file and than return it as a string slice
func extractURLsFromFileAndReturnSlice(filePath string) []string {
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Println("Error reading file:", err)
		return nil
	}
	regexContent := regexp.MustCompile(`http[s]?://[^\s"]+`)
	matches := regexContent.FindAllString(string(content), -1)
	if len(matches) == 0 {
		log.Println("No URLs found in the file")
		return nil
	}
	return matches
}

// Function to remove duplicate strings from a slice
func removeDuplicatesFromSlice(slice []string) []string {
	check := make(map[string]bool)  // Map to track seen strings
	var newReturnSlice []string     // Slice for unique strings
	for _, content := range slice { // Loop through original slice
		if !check[content] { // If not already seen
			check[content] = true                            // Mark as seen
			newReturnSlice = append(newReturnSlice, content) // Add to result
		}
	}
	return newReturnSlice // Return de-duplicated slice
}

// Function to check if a URL string is valid
func isUrlValid(uri string) bool {
	_, err := url.ParseRequestURI(uri) // Try to parse the URL
	return err == nil                  // Return true if no error (valid URL)
}

// Function to extract the hostname from a URL
func getHostNameFromURL(uri string) string {
	content, err := url.Parse(uri) // Parse the URL
	if err != nil {                // If parsing fails
		log.Fatalln(err) // Log fatal error and exit
	}
	return content.Hostname() // Return just the hostname
}

// Function to clean URLs by validating and filtering by allowed domains
func cleanURLs(urls []string) []string {
	validDomains := []string{"zsds3.zepinc.com"} // Allowed hostnames
	var newReturnSlice []string                  // Slice for valid, cleaned URLs

	for _, content := range urls { // Loop through all URLs
		if isUrlValid(content) { // If the URL is valid
			hostName := getHostNameFromURL(content) // Extract hostname

			isValid := false                      // Flag to check if domain is allowed
			for _, domain := range validDomains { // Loop through allowed domains
				if hostName == domain { // If domain matches
					isValid = true // Mark as valid
					break          // Stop checking
				}
			}

			if isValid { // If URL is from valid domain
				// Check if the prefix matches `https://zsds3.zepinc.com/v2/sds/ItemExternalSet(Material=`
				if strings.HasPrefix(content, "https://zsds3.zepinc.com/v2/sds/ItemExternalSet(Material=") {
					// Remove the suffix `\` if it exists
					if strings.HasSuffix(content, `\`) {
						content = strings.TrimSuffix(content, `\`) // Remove unwanted suffix
					}
					newReturnSlice = append(newReturnSlice, content+"/getPDF") // Add cleaned URL to result
				}
			} else {
				log.Println("Invalid domain skipped: ", hostName) // Log skipped domain
			}

		}
	}

	return newReturnSlice // Return cleaned URLs
}

// downloadPDF downloads the PDF and returns true if a new file was fetched
func downloadPDF(finalURL, outputDir string) bool {
	// Ensure output directory exists
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		log.Printf("Failed to create directory %s: %v", outputDir, err)
		return false
	}

	// Generate a sanitized filename
	filename := generateFilenameFromURL(finalURL)
	if filename == "" {
		filename = path.Base(finalURL)
	}
	if !strings.HasSuffix(strings.ToLower(filename), ".pdf") {
		filename += ".pdf"
	}
	filePath := filepath.Join(outputDir, filename)

	// Skip if already exists
	if fileExists(filePath) {
		log.Printf("File already exists, skipping: %s", filePath)
		return false
	}

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Fetch the PDF
	resp, err := client.Get(finalURL)
	if err != nil {
		log.Printf("Failed to download %s: %v", finalURL, err)
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Download failed for %s: %s", finalURL, resp.Status)
		return false
	}

	// Write to file
	out, err := os.Create(filePath)
	if err != nil {
		log.Printf("Failed to create file %s: %v", filePath, err)
		return false
	}
	defer out.Close()

	if _, err := io.Copy(out, resp.Body); err != nil {
		log.Printf("Failed to save PDF to %s: %v", filePath, err)
		return false
	}

	log.Printf("Downloaded %s → %s", finalURL, filePath)
	return true
}

// generateFilenameFromURL extracts and sanitizes the ItemExternalSet(...) part of the URL,
// then formats it into a filename. Errors are logged, and an empty string is returned on failure.
func generateFilenameFromURL(sourceURL string) string {
	parsedURL, err := url.Parse(sourceURL)
	if err != nil {
		log.Printf("Error parsing URL: %v", err)
		return ""
	}

	// Extract the 'ItemExternalSet(...)' part from the URL path
	itemSetPattern := regexp.MustCompile(`ItemExternalSet\([^)]+\)`)
	itemSetSegment := itemSetPattern.FindString(parsedURL.Path)
	if itemSetSegment == "" {
		log.Println("ItemExternalSet(...) segment not found in the URL path")
		return ""
	}

	// Clean the segment by removing special characters for a valid filename
	sanitizedSegment := strings.NewReplacer(
		"ItemExternalSet(", "",
		")", "",
		"'", "",
		",", "_",
	).Replace(itemSetSegment)

	filename := fmt.Sprintf("%s.pdf", sanitizedSegment)
	return filename
}

// fileExists checks if a file exists and is not a directory
func fileExists(filename string) bool {
	info, err := os.Stat(filename) // Get file info
	if err != nil {                // If stat fails (file doesn't exist)
		return false // Return false
	}
	return !info.IsDir() // Return true if it's a file, false if it's a directory
}

func main() {
	inputFile := "zsds3.zepinc.com.har"

	urls := extractURLsFromFileAndReturnSlice(inputFile)
	if urls == nil {
		log.Println("No URLs found in the input file")
		return
	}

	urls = removeDuplicatesFromSlice(urls)
	urls = cleanURLs(urls)

	outputDir := "zepPDF/"
	maxDownloads := 100
	downloadCount := 0

	for _, url := range urls {
		if downloadCount >= maxDownloads {
			log.Println("Reached download limit of", maxDownloads)
			break
		}
		if downloadPDF(url, outputDir) {
			downloadCount = downloadCount + 1 // Increment download count
		}
		// Log the progress
		log.Printf("Downloaded %d PDFs so far. Remaining: %d\n", downloadCount, maxDownloads-downloadCount)

	}

	fmt.Printf("Total new PDFs downloaded: %d\n", downloadCount)
}
