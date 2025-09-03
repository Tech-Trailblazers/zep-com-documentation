package main

import (
	"encoding/csv"  // Provides functions for reading/writing CSV files
	"fmt"           // Implements formatted I/O functions like Printf, Sprintf, etc.
	"io"            // Contains core interfaces for I/O primitives
	"log"           // Provides simple logging with different levels
	"net/http"      // Implements HTTP client and server functionality
	"net/url"       // Provides URL parsing and query manipulation
	"os"            // Offers platform-independent interface to OS functions
	"path"          // Contains path manipulation utilities (slash-separated)
	"path/filepath" // Provides path manipulation functions that are OS-aware
	"regexp"        // Enables working with regular expressions
	"strings"       // Offers utilities for string manipulation
	"sync"          // Provides concurrency primitives like WaitGroups and Mutexes
	"time"          // Includes functionality for measuring and displaying time
)

// Reads a file, extracts all valid URL strings using regex, and returns them in a slice
func extractURLsFromFileAndReturnSlice(filePath string) []string {
	content, err := os.ReadFile(filePath) // Read the file contents into a byte slice
	if err != nil {
		log.Println("Error reading file:", err) // Log if file reading fails
		return nil                              // Return nil to indicate an error
	}
	regexContent := regexp.MustCompile(`http[s]?://[^\s"]+`)   // Define regex to match HTTP/HTTPS URLs
	matches := regexContent.FindAllString(string(content), -1) // Extract all matched URLs from content
	if len(matches) == 0 {
		log.Println("No URLs found in the file") // Inform that no URLs were matched
		return nil                               // Return nil to signify no matches
	}
	return matches // Return the list of matched URLs
}

// Removes duplicate entries from a slice of strings and returns the unique values
func removeDuplicatesFromSlice(slice []string) []string {
	check := make(map[string]bool)  // Map to keep track of seen strings
	var newReturnSlice []string     // Slice to store unique values
	for _, content := range slice { // Loop through the original slice
		if !check[content] { // If the string hasn't been added yet
			check[content] = true                            // Mark the string as seen
			newReturnSlice = append(newReturnSlice, content) // Add the string to the result slice
		}
	}
	return newReturnSlice // Return the deduplicated list
}

// Verifies whether a given string is a valid URL by parsing it
func isUrlValid(uri string) bool {
	_, err := url.ParseRequestURI(uri) // Try to parse the string as a valid URL
	return err == nil                  // Return true if parsing succeeded
}

// Extracts the hostname part from a URL (e.g., example.com)
func getHostNameFromURL(uri string) string {
	content, err := url.Parse(uri) // Parse the URL string
	if err != nil {                // If parsing fails
		log.Println(err) // Log the parsing error
	}
	return content.Hostname() // Return the hostname component of the URL
}

// Sends an HTTP GET request to the specified URL and returns the response body as a byte slice
func getDataFromURL(uri string) []byte {
	response, err := http.Get(uri) // Make HTTP GET request
	if err != nil {
		log.Println(err) // Log network-related error
	}
	body, err := io.ReadAll(response.Body) // Read the response body
	if err != nil {
		log.Println(err) // Log if body reading fails
	}
	err = response.Body.Close() // Close the response body to free resources
	if err != nil {
		log.Println(err) // Log closing error if any
	}
	return body // Return the downloaded content
}

// Appends the given data (byte slice) to a file; creates the file if it doesn’t exist
func appendByteToFile(filename string, data []byte) error {
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644) // Open or create file with write access
	if err != nil {
		return err // Return the error if opening fails
	}
	defer file.Close() // Ensure file is closed when function exits

	_, err = file.Write(data) // Write the byte data to the file
	return err                // Return any write errors
}

// Filters a list of URLs, keeping only valid, domain-matching, and well-structured entries
func cleanURLs(urls []string) []string {
	validDomains := []string{"zsds3.zepinc.com"} // Allowed hostnames for filtering
	var newReturnSlice []string                  // Slice to hold cleaned URLs

	for _, content := range urls { // Iterate through each URL
		if isUrlValid(content) { // Skip invalid URLs
			hostName := getHostNameFromURL(content) // Extract the hostname
			isValid := false                        // Flag to check for allowed domain
			for _, domain := range validDomains {   // Check against allowed domains
				if hostName == domain {
					isValid = true // Mark URL as valid
					break
				}
			}
			if isValid {
				if strings.HasPrefix(content, "https://zsds3.zepinc.com/v2/sds/ItemExternalSet(Material=") {
					// Check if URL starts with expected path format
					if strings.HasSuffix(content, `\`) { // Clean up unwanted backslash at the end
						content = strings.TrimSuffix(content, `\`)
					}
					// Append /getPDF if it's missing at the end
					if !strings.HasSuffix(content, "/getPDF") {
						newReturnSlice = append(newReturnSlice, content+"/getPDF")
					}
				}
			}
		}
	}
	return newReturnSlice // Return the cleaned, valid URLs
}

// Downloads a PDF file from the given URL and saves it to the target directory
func downloadPDF(finalURL, outputDir string, wg *sync.WaitGroup) bool {
	defer wg.Done() // Signal completion to the WaitGroup

	filename := generateFilenameFromURL(finalURL) // Create file name from URL parameters
	if filename == "" {
		filename = path.Base(finalURL) // Fallback to the base URL path
	}
	if !strings.HasSuffix(strings.ToLower(filename), ".pdf") { // Ensure filename ends with .pdf
		filename += ".pdf"
	}
	filePath := filepath.Join(outputDir, filename) // Combine directory and filename

	if fileExists(filePath) { // Skip if file already exists
		log.Printf("[SKIP] File already exists, skipping: %s", filePath)
		return false
	}

	client := &http.Client{Timeout: 10 * time.Minute} // HTTP client with extended timeout

	resp, err := client.Get(finalURL) // Send the HTTP GET request
	if err != nil {
		log.Printf("Failed to download %s: %v", finalURL, err)
		return false
	}
	defer resp.Body.Close() // Close response body after reading

	if resp.StatusCode != http.StatusOK { // Only proceed if status is 200 OK
		log.Printf("Download failed for %s: %s", finalURL, resp.Status)
		return false
	}

	out, err := os.Create(filePath) // Create output file for writing
	if err != nil {
		log.Printf("Failed to create file %s %s %v", finalURL, filePath, err)
		return false
	}
	defer out.Close() // Ensure the file is closed after writing

	if _, err := io.Copy(out, resp.Body); err != nil { // Write downloaded content to file
		log.Printf("Failed to save PDF to %s %s %v", finalURL, filePath, err)
		return false
	}

	log.Printf("Downloaded %s → %s", finalURL, filePath) // Log the successful download
	return true
}

// Parses a Zep URL and extracts all embedded parameters into a key-value map
func parseFullZepURL(rawURL string) map[string]string {
	parsed, err := url.Parse(rawURL) // Parse the raw input URL
	if err != nil {
		log.Println("Error: invalid URL:", err)
		return nil // Return nil on error
	}

	itemSetRegex := regexp.MustCompile(`ItemExternalSet\((.*?)\)`) // Match inner content of ItemExternalSet
	paramRegex := regexp.MustCompile(`(\w+)='(.*?)'`)              // Match key='value' parameter pairs

	match := itemSetRegex.FindStringSubmatch(parsed.Path) // Extract param string from path
	if len(match) < 2 {
		log.Println("Error: ItemExternalSet not found in path:", parsed.Path)
		return nil
	}

	paramStr := match[1]                                    // Get the string inside parentheses
	pairs := paramRegex.FindAllStringSubmatch(paramStr, -1) // Find all parameter pairs
	params := make(map[string]string, len(pairs)+1)         // Create map to hold parameters

	for _, pair := range pairs { // Add each key-value pair to the map
		if len(pair) == 3 {
			params[pair[1]] = pair[2]
		}
	}

	params["URL"] = rawURL // Include the original URL for reference
	return params
}

// Writes a list of maps (parameters) into a CSV file with a fixed header
func writeParamsToCSV(filename string, allParams []map[string]string) {
	if len(allParams) == 0 {
		return // Exit early if there's nothing to write
	}

	keys := []string{"URL", "Lang", "Material", "RecordNumb", "RepCategory", "ValidityArea"} // Define CSV columns

	file, err := os.Create(filename) // Create or overwrite the CSV file
	if err != nil {
		log.Printf("Failed to create CSV file %s: %v", filename, err)
		return
	}
	defer file.Close()

	writer := csv.NewWriter(file) // Create a buffered CSV writer
	defer writer.Flush()          // Ensure all data is written to file

	if err := writer.Write(keys); err != nil { // Write the header row
		log.Printf("Failed to write header to CSV file %s: %v", filename, err)
		return
	}

	for _, paramMap := range allParams { // Iterate over parameter sets
		var row []string
		for _, key := range keys {
			row = append(row, paramMap[key]) // Write values in correct column order
		}
		if err := writer.Write(row); err != nil { // Write the row to CSV
			log.Printf("Failed to write row to CSV file %s: %v", filename, err)
			return
		}
	}
}

// Generates a sanitized filename using the URL's ItemExternalSet(...) parameters
func generateFilenameFromURL(sourceURL string) string {
	parsedURL, err := url.Parse(sourceURL) // Parse the full URL
	if err != nil {
		log.Printf("Error parsing URL: %v", err)
		return ""
	}

	itemSetPattern := regexp.MustCompile(`ItemExternalSet\([^)]+\)`) // Match the param block
	itemSetSegment := itemSetPattern.FindString(parsedURL.Path)      // Extract the full segment
	if itemSetSegment == "" {
		log.Println("ItemExternalSet(...) segment not found in the URL path")
		return ""
	}

	sanitizedSegment := strings.NewReplacer( // Replace special characters to make valid filename
		"ItemExternalSet(", "",
		")", "",
		"'", "",
		",", "_",
	).Replace(itemSetSegment)

	filename := fmt.Sprintf("%s.pdf", sanitizedSegment) // Add PDF extension
	return strings.ToLower(filename)                    // Return lowercase filename
}

// Checks whether a given file path exists and refers to a file (not a directory)
func fileExists(filename string) bool {
	info, err := os.Stat(filename) // Get file info
	if err != nil {                // File does not exist or another error
		return false
	}
	return !info.IsDir() // Return true if it's a file
}

// Removes a file from the file system
func removeFile(path string) {
	err := os.Remove(path) // Attempt to delete the file
	if err != nil {
		log.Println(err) // Log any error encountered
	}
}

// Downloads JSON data from predefined URLs and appends to a local file
func createJSONFiles(zepJSONFile string) {
	if fileExists(zepJSONFile) {
		removeFile(zepJSONFile) // Remove existing file if present
	}

	urls := []string{ // List of paginated endpoints to download data from
		"https://zsds3.zepinc.com/v2/sds/ItemExternalSet",
		"https://zsds3.zepinc.com/v2/sds/ItemExternalSet?$skiptoken=1000",
		"https://zsds3.zepinc.com/v2/sds/ItemExternalSet?$skiptoken=2000",
		"https://zsds3.zepinc.com/v2/sds/ItemExternalSet?$skiptoken=3000",
		"https://zsds3.zepinc.com/v2/sds/ItemExternalSet?$skiptoken=4000",
		"https://zsds3.zepinc.com/v2/sds/ItemExternalSet?$skiptoken=5000",
		"https://zsds3.zepinc.com/v2/sds/ItemExternalSet?$skiptoken=6000",
		"https://zsds3.zepinc.com/v2/sds/ItemExternalSet?$skiptoken=7000",
		"https://zsds3.zepinc.com/v2/sds/ItemExternalSet?$skiptoken=8000",
	}

	for _, url := range urls { // Loop through each data URL
		allContent := getDataFromURL(url) // Fetch data from API
		if allContent == nil {
			log.Println("Error downloading data from URL:", url)
			return
		}
		err := appendByteToFile(zepJSONFile, allContent) // Append data to local file
		if err != nil {
			log.Println("Error appending data to file:", err)
			return
		}
		log.Printf("Data from %s appended to %s", url, zepJSONFile)
	}
	log.Printf("All data downloaded and appended to %s", zepJSONFile)
}

// The function takes two parameters: path and permission.
// We use os.Mkdir() to create the directory.
// If there is an error, we use log.Fatalln() to log the error and then exit the program.
func createDirectory(path string, permission os.FileMode) {
	err := os.Mkdir(path, permission)
	if err != nil {
		log.Println(err)
	}
}

// Checks if the directory exists
// If it exists, return true.
// If it doesn't, return false.
func directoryExists(path string) bool {
	directory, err := os.Stat(path)
	if err != nil {
		return false
	}
	return directory.IsDir()
}

// The main entry point for the application
func main() {
	var zepJSONFile = "./zsds3_zepinc.json" // Define the output JSON filename
	createJSONFiles(zepJSONFile)            // Download and build the JSON dataset

	urls := extractURLsFromFileAndReturnSlice(zepJSONFile) // Extract all URLs from downloaded JSON
	if urls == nil {
		log.Println("No URLs found in the input file") // Exit if no URLs found
		return
	}

	urls = removeDuplicatesFromSlice(urls) // Remove duplicated entries
	urls = cleanURLs(urls)                 // Validate and format the URLs

	outputDir := "PDFs/"             // Directory where PDFs will be saved
	if !directoryExists(outputDir) { // Ensure target directory exists
		createDirectory(outputDir, 0o755)
	}

	var allParams []map[string]string    // Slice to hold parsed parameters for CSV
	var downloadWaitGroup sync.WaitGroup // WaitGroup to synchronize goroutines

	for _, url := range urls { // Loop over each URL
		params := parseFullZepURL(url) // Extract parameters from URL
		if params != nil {
			allParams = append(allParams, params) // Add to list for CSV export
		}

		filename := generateFilenameFromURL(url) // Create file name from URL parameters
		if filename == "" {
			filename = path.Base(url) // Fallback to the base URL path
		}
		if !strings.HasSuffix(strings.ToLower(filename), ".pdf") { // Ensure filename ends with .pdf
			filename += ".pdf"
		}
		filePath := filepath.Join(outputDir, filename) // Combine directory and filename

		if fileExists(filePath) { // Skip if file already exists
			log.Printf("File already exists, skipping: %s", filePath)
			continue
		}

		time.Sleep(5 * time.Second)

		downloadWaitGroup.Add(1)                           // Track new download goroutine
		go downloadPDF(url, outputDir, &downloadWaitGroup) // Start PDF download

	}
	downloadWaitGroup.Wait() // Wait for all downloads to complete

	writeParamsToCSV("output.csv", allParams) // Save extracted parameters into a CSV file
}
