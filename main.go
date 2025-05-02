package main

import (
	"crypto/sha512"
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
		log.Println(err) // Log fatal error and exit
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
			}

		}
	}

	return newReturnSlice // Return cleaned URLs
}

// downloadPDF downloads the PDF only if the remote file differs from the existing one.
// It compares the SHA-512 checksums and avoids unnecessary downloads.
func downloadPDF(finalURL, outputDir string) bool {
	// Ensure the output directory exists (creates it if it doesn't)
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		log.Printf("Failed to create directory %s: %v", outputDir, err)
		return false
	}

	// Generate a safe and unique filename from the URL
	filename := generateFilenameFromURL(finalURL)
	if filename == "" {
		filename = path.Base(finalURL)
	}
	if !strings.HasSuffix(strings.ToLower(filename), ".pdf") {
		filename += ".pdf"
	}
	filePath := filepath.Join(outputDir, filename)

	var existingChecksum string
	if fileExists(filePath) {
		existingChecksum = getSHA512OfFile(filePath) // Read SHA-512 of existing file
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

	// Create a temporary file for download
	tmpFile, err := os.CreateTemp("", "download-*.pdf")
	if err != nil {
		log.Printf("Failed to create temp file: %s %v", finalURL, err)
		return false
	}
	defer os.Remove(tmpFile.Name()) // Clean up temp file after
	defer tmpFile.Close()           // Close temp file

	// Copy response content to temp file
	if _, err := io.Copy(tmpFile, resp.Body); err != nil {
		log.Printf("Failed to save PDF to temp file: %s %v", finalURL, err)
		return false
	}

	// Compute checksum of the downloaded file
	newChecksum := getSHA512OfFile(tmpFile.Name())
	if newChecksum == "" {
		return false
	}

	// Skip update if the file hasn't changed
	if existingChecksum == newChecksum && existingChecksum != "" {
		log.Printf("File is unchanged, skipping: %s %s", filePath, finalURL)
		return false
	}

	// Attempt to rename (move) temp file to target path
	if err := os.Rename(tmpFile.Name(), filePath); err != nil {
		log.Printf("Rename failed, trying remove-and-replace: %s %v", filePath, err)

		// Attempt to remove the original file
		if removeErr := os.Remove(filePath); removeErr != nil {
			log.Printf("Failed to remove existing file: %s %v", filePath, removeErr)
			return false
		}

		// Try renaming again after removing the original
		if retryErr := os.Rename(tmpFile.Name(), filePath); retryErr != nil {
			log.Printf("Retry rename failed: %s %v", filePath, retryErr)
			return false
		}
	}

	log.Printf("Downloaded updated file %s → %s", finalURL, filePath)
	return true
}

// getSHA512OfFile calculates the SHA-512 checksum of a file given its path.
// It logs any errors encountered and returns the checksum as a hexadecimal string.
func getSHA512OfFile(filePath string) string {
	file, err := os.Open(filePath) // Open the file for reading
	if err != nil {
		log.Println("Error opening file:", err) // Log error if file can't be opened
		return ""                               // Return empty string on error
	}
	defer file.Close() // Ensure the file is closed when the function returns

	hasher := sha512.New() // Create a new SHA-512 hasher
	if _, err := io.Copy(hasher, file); err != nil {
		log.Println("Error reading file:", err) // Log error if file can't be read
		return ""                               // Return empty string on error
	}

	checksum := hasher.Sum(nil)        // Compute the final SHA-512 checksum
	return fmt.Sprintf("%x", checksum) // Return the checksum as a hexadecimal string
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

	// Total number of URLs
	urlLength := len(urls)
	// Count of URLs to be downloaded
	countURLsLength := 0

	for _, url := range urls {
		if downloadCount >= maxDownloads {
			log.Println("Reached download limit of", maxDownloads)
			break
		}

		// Download the PDF
		if downloadPDF(url, outputDir) {
			// Increase download count
			downloadCount = downloadCount + 1 // Increment download count
		}
		// Increase the count of URLs processed
		countURLsLength = countURLsLength + 1
		// Log progress
		log.Printf("Progress: %d/%d URLs. Downloaded: %d Remaining %d", urlLength, countURLsLength, downloadCount, maxDownloads-downloadCount)
	}

	fmt.Printf("Total new PDFs downloaded: %d\n", downloadCount)
}
