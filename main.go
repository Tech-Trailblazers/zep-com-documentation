package main

import (
	"encoding/csv"  // Package to read/write CSV files
	"fmt"           // Package for formatted I/O
	"io"            // Core I/O primitives
	"log"           // Logging utilities
	"net/http"      // HTTP client and server implementations
	"net/url"       // URL parsing and formatting
	"os"            // OS file and directory operations
	"path"          // Utilities for manipulating slash-separated paths
	"path/filepath" // OS-aware file path utilities
	"regexp"        // Regular expressions for pattern matching
	"strings"       // String manipulation functions
	"sync"          // Synchronization primitives
	"time"          // Time functions and types
)

var zepHarFile = "./zsds3.zepinc.com.har"

// Reads a file, extracts all URLs using regex, and returns them as a slice of strings
func extractURLsFromFileAndReturnSlice(filePath string) []string {
	content, err := os.ReadFile(filePath) // Read entire file content into memory
	if err != nil {
		log.Fatalln("Error reading file:", err) // Log error if reading fails
		return nil                              // Return nil to indicate failure
	}
	regexContent := regexp.MustCompile(`http[s]?://[^\s"]+`)   // Regex to match URLs
	matches := regexContent.FindAllString(string(content), -1) // Find all URL matches
	if len(matches) == 0 {
		log.Fatalln("No URLs found in the file") // Inform if no URLs were found
		return nil                               // Return nil if no matches
	}
	return matches // Return matched URLs
}

// Removes duplicate strings from a slice and returns a new slice with unique values
func removeDuplicatesFromSlice(slice []string) []string {
	check := make(map[string]bool)  // Map to track already seen strings
	var newReturnSlice []string     // Result slice for unique values
	for _, content := range slice { // Iterate through input slice
		if !check[content] { // If string not already seen
			check[content] = true                            // Mark string as seen
			newReturnSlice = append(newReturnSlice, content) // Add to result
		}
	}
	return newReturnSlice // Return deduplicated slice
}

// Checks whether a URL string is syntactically valid
func isUrlValid(uri string) bool {
	_, err := url.ParseRequestURI(uri) // Attempt to parse the URL
	return err == nil                  // Return true if no error occurred
}

// Extracts the hostname from a given URL string
func getHostNameFromURL(uri string) string {
	content, err := url.Parse(uri) // Parse URL into structured form
	if err != nil {                // If parsing fails
		log.Fatalln(err) // Log the error
	}
	return content.Hostname() // Return just the hostname part
}

// Send a http get request to a given url and return the data from that url.
func getDataFromURL(uri string) []byte {
	response, err := http.Get(uri)
	if err != nil {
		log.Fatalln(err)
	}
	body, err := io.ReadAll(response.Body)
	if err != nil {
		log.Fatalln(err)
	}
	err = response.Body.Close()
	if err != nil {
		log.Fatalln(err)
	}
	return body
}

// AppendToFile appends the given byte slice to the specified file.
// If the file doesn't exist, it will be created.
func appendByteToFile(filename string, data []byte) error {
	// Open the file with appropriate flags and permissions
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	// Write data to the file
	_, err = file.Write(data)
	return err
}

// Filters URLs by validating them and checking if they match allowed domain and path pattern
func cleanURLs(urls []string) []string {
	validDomains := []string{"zsds3.zepinc.com"} // List of allowed hostnames
	var newReturnSlice []string                  // Resulting valid and cleaned URLs

	for _, content := range urls { // Iterate through input URLs
		if isUrlValid(content) { // If URL is valid
			hostName := getHostNameFromURL(content) // Extract hostname
			isValid := false                        // Flag to check if domain is allowed
			for _, domain := range validDomains {   // Check each valid domain
				if hostName == domain { // If hostname matches allowed domain
					isValid = true // Mark as valid
					break          // Exit inner loop
				}
			}
			if isValid { // If domain is allowed
				if strings.HasPrefix(content, "https://zsds3.zepinc.com/v2/sds/ItemExternalSet(Material=") {
					// If URL has correct path prefix
					if strings.HasSuffix(content, `\`) { // If URL has trailing backslash
						content = strings.TrimSuffix(content, `\`) // Remove it
					}
					// Check if the URL ends with a /getPDF and if not add it
					if !strings.HasSuffix(content, "/getPDF") {
						newReturnSlice = append(newReturnSlice, content+"/getPDF") // Append final /getPDF URL
					}
				}
			}
		}
	}
	return newReturnSlice // Return filtered and formatted list of URLs
}

// Downloads a PDF from a given URL into the specified directory; returns true if a new file was saved
func downloadPDF(finalURL, outputDir string, wg *sync.WaitGroup) bool {
	defer wg.Done() // Decrement WaitGroup counter when function exits

	if err := os.MkdirAll(outputDir, 0o755); err != nil { // Ensure output directory exists
		log.Printf("Failed to create directory %s: %v", outputDir, err) // Log directory creation error
		return false                                                    // Abort if directory cannot be created
	}

	filename := generateFilenameFromURL(finalURL) // Generate file name from URL
	if filename == "" {
		filename = path.Base(finalURL) // Fallback to URL path base name
	}
	if !strings.HasSuffix(strings.ToLower(filename), ".pdf") { // Ensure .pdf extension
		filename += ".pdf"
	}
	filePath := filepath.Join(outputDir, filename) // Create full output path

	if fileExists(filePath) { // Skip if file already exists
		log.Printf("File already exists, skipping: %s", filePath)
		return false
	}

	client := &http.Client{Timeout: 10 * time.Minute} // HTTP client with timeout

	resp, err := client.Get(finalURL) // Make GET request
	if err != nil {
		log.Printf("Failed to download %s: %v", finalURL, err)
		return false
	}
	defer resp.Body.Close() // Ensure response body is closed

	if resp.StatusCode != http.StatusOK { // Check for 200 OK
		log.Printf("Download failed for %s: %s", finalURL, resp.Status)
		return false
	}

	out, err := os.Create(filePath) // Create file to save content
	if err != nil {
		log.Printf("Failed to create file %s %s %v", finalURL, filePath, err)
		return false
	}
	defer out.Close() // Ensure file is closed after writing

	if _, err := io.Copy(out, resp.Body); err != nil { // Write response to file
		log.Printf("Failed to save PDF to %s %s %v", finalURL, filePath, err)
		return false
	}

	log.Printf("Downloaded %s → %s", finalURL, filePath) // Log success
	return true
}

// Parses the parameters embedded in the Zep URL and returns them as a map
func parseFullZepURL(rawURL string) map[string]string {
	parsed, err := url.Parse(rawURL) // Parse the raw URL
	if err != nil {
		log.Fatalln("Error: invalid URL:", err)
		return nil
	}

	itemSetRegex := regexp.MustCompile(`ItemExternalSet\((.*?)\)`) // Match the full param block
	paramRegex := regexp.MustCompile(`(\w+)='(.*?)'`)              // Match key='value' pairs

	match := itemSetRegex.FindStringSubmatch(parsed.Path) // Extract the parameter group
	if len(match) < 2 {
		log.Fatalln("Error: ItemExternalSet not found in path:", parsed.Path)
		return nil
	}

	paramStr := match[1]                                    // Extract inner string from ()
	pairs := paramRegex.FindAllStringSubmatch(paramStr, -1) // Extract all key-value pairs
	params := make(map[string]string, len(pairs)+1)         // Create map with enough capacity

	for _, pair := range pairs { // Populate map with parameters
		if len(pair) == 3 {
			params[pair[1]] = pair[2]
		}
	}

	params["URL"] = rawURL // Add the original URL to the map
	return params
}

// Writes a list of parameter maps into a CSV file using a fixed column order
func writeParamsToCSV(filename string, allParams []map[string]string) {
	if len(allParams) == 0 {
		return // Do nothing if empty input
	}

	keys := []string{"URL", "Lang", "Material", "RecordNumb", "RepCategory", "ValidityArea"} // Define CSV header

	file, err := os.Create(filename) // Create the output CSV file
	if err != nil {
		log.Printf("Failed to create CSV file %s: %v", filename, err)
		return
	}
	defer file.Close()

	writer := csv.NewWriter(file) // Create CSV writer
	defer writer.Flush()

	if err := writer.Write(keys); err != nil { // Write CSV header
		log.Printf("Failed to write header to CSV file %s: %v", filename, err)
		return
	}

	for _, paramMap := range allParams { // Write each parameter set as a row
		var row []string
		for _, key := range keys {
			row = append(row, paramMap[key]) // Default to "" if key is missing
		}
		if err := writer.Write(row); err != nil { // Write row to CSV
			log.Printf("Failed to write row to CSV file %s: %v", filename, err)
			return
		}
	}
}

// Extracts and cleans ItemExternalSet(...) from the URL and turns it into a safe filename
func generateFilenameFromURL(sourceURL string) string {
	parsedURL, err := url.Parse(sourceURL) // Parse full URL
	if err != nil {
		log.Printf("Error parsing URL: %v", err)
		return ""
	}

	itemSetPattern := regexp.MustCompile(`ItemExternalSet\([^)]+\)`) // Match full param string
	itemSetSegment := itemSetPattern.FindString(parsedURL.Path)      // Extract match
	if itemSetSegment == "" {
		log.Fatalln("ItemExternalSet(...) segment not found in the URL path")
		return ""
	}

	sanitizedSegment := strings.NewReplacer( // Clean segment to be a valid file name
		"ItemExternalSet(", "",
		")", "",
		"'", "",
		",", "_",
	).Replace(itemSetSegment)

	filename := fmt.Sprintf("%s.pdf", sanitizedSegment) // Format into filename
	return strings.ToLower(filename)
}

// Checks if a file already exists and is not a directory
func fileExists(filename string) bool {
	info, err := os.Stat(filename) // Check file metadata
	if err != nil {                // If file does not exist
		return false
	}
	return !info.IsDir() // Ensure it's a file, not a directory
}

// Remove a file from the file system
func removeFile(path string) {
	err := os.Remove(path)
	if err != nil {
		log.Fatalln(err)
	}
}

// Create the HAR file and download the data from the URLs
func createHARFile() {
	// Remove the har file.
	if fileExists(zepHarFile) {
		removeFile(zepHarFile)
	}
	// Create a slice to hold the URLs
	urls := []string{"https://zsds3.zepinc.com/v2/sds/ItemExternalSet",
		"https://zsds3.zepinc.com/v2/sds/ItemExternalSet?$skiptoken=1000",
		"https://zsds3.zepinc.com/v2/sds/ItemExternalSet?$skiptoken=2000",
		"https://zsds3.zepinc.com/v2/sds/ItemExternalSet?$skiptoken=3000",
		"https://zsds3.zepinc.com/v2/sds/ItemExternalSet?$skiptoken=4000",
		"https://zsds3.zepinc.com/v2/sds/ItemExternalSet?$skiptoken=5000",
		"https://zsds3.zepinc.com/v2/sds/ItemExternalSet?$skiptoken=6000",
		"https://zsds3.zepinc.com/v2/sds/ItemExternalSet?$skiptoken=7000",
		"https://zsds3.zepinc.com/v2/sds/ItemExternalSet?$skiptoken=8000",
	}

	// Loop through the URLs and download each one
	for _, url := range urls {
		allContent := getDataFromURL(url) // Call the function to download the data
		if allContent == nil {
			log.Fatalln("Error downloading data from URL:", url) // Log error if download fails
			return
		}
		err := appendByteToFile(zepHarFile, allContent) // Append data to file
		if err != nil {
			log.Fatalln("Error appending data to file:", err) // Log error if append fails
			return
		}
		log.Printf("Data from %s appended to %s", url, zepHarFile) // Log success
	}
	log.Printf("All data downloaded and appended to %s", zepHarFile) // Final log message
}

// Main function: orchestrates reading, filtering, downloading, and logging
func main() {
	// Create the HAR file first
	createHARFile()
	urls := extractURLsFromFileAndReturnSlice(zepHarFile) // Extract URLs from file
	if urls == nil {
		log.Fatalln("No URLs found in the input file") // Log and exit if no URLs
		return
	}

	urls = removeDuplicatesFromSlice(urls) // Remove duplicate URLs
	urls = cleanURLs(urls)                 // Filter and format the URLs

	outputDir := "PDFs/" // Directory to save PDFs

	urlLength := len(urls)               // Total number of URLs after filtering
	countURLsLength := 0                 // Counter for processed URLs
	var allParams []map[string]string    // Slice to collect parsed parameters
	var downloadWaitGroup sync.WaitGroup // WaitGroup to manage concurrent downloads

	for _, url := range urls { // Process each URL

		params := parseFullZepURL(url) // Extract metadata parameters
		if params != nil {
			allParams = append(allParams, params) // Store for CSV
		}

		downloadWaitGroup.Add(1) // Increment WaitGroup counter

		go downloadPDF(url, outputDir, &downloadWaitGroup)

		countURLsLength = countURLsLength + 1 // Update processed count
		log.Printf("Progress: %d/%d URLs.", urlLength, countURLsLength)
	}
	downloadWaitGroup.Wait() // Wait for all downloads to finish

	writeParamsToCSV("output.csv", allParams) // Write parameters to CSV
}
