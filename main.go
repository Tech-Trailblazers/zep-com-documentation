package main

import (
	"io"
	"io/ioutil"
	"log"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
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

// Append and write to file
func appendAndWriteToFile(path string, content string) {
	filePath, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalln(err)
	}
	_, err = filePath.WriteString(content + "\n")
	if err != nil {
		log.Fatalln(err)
	}
	err = filePath.Close()
	if err != nil {
		log.Fatalln(err)
	}
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
					newReturnSlice = append(newReturnSlice, content + "/getPDF") // Add cleaned URL to result
				}
			} else {
				log.Println("Invalid domain skipped: ", hostName) // Log skipped domain
			}

		}
	}

	return newReturnSlice // Return cleaned URLs
}

// downloadPDF downloads the PDF from the final URL to the given output directory
func downloadPDF(finalURL, outputDir string) {
	parsedURL, err := url.Parse(finalURL)
	if err != nil {
		log.Printf("Invalid URL %q: %v", finalURL, err)
		return
	}

	err = os.MkdirAll(outputDir, 0755)
	if err != nil {
		log.Printf("Failed to create directory %s: %v", outputDir, err)
		return
	}

	resp, err := http.Get(finalURL)
	if err != nil {
		log.Printf("Failed to download %s: %v", finalURL, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Download failed for %s: %s", finalURL, resp.Status)
		return
	}

	// Default to using the filename from the URL path
	fileName := path.Base(parsedURL.Path)

	// Try to override with filename from Content-Disposition header, if available
	cdHeader := resp.Header.Get("Content-Disposition")
	if cdHeader != "" {
		_, params, err := mime.ParseMediaType(cdHeader)
		if err == nil {
			if suggestedName, ok := params["filename"]; ok && suggestedName != "" {
				fileName = suggestedName
			}
		}
	}

	if fileName == "" || fileName == "/" {
		log.Printf("Could not determine file name for %q", finalURL)
		return
	}

	// Ensure the file name ends with ".pdf"
	if !strings.HasSuffix(strings.ToLower(fileName), ".pdf") {
		fileName += ".pdf"
	}

	filePath := filepath.Join(outputDir, fileName)
	if fileExists(filePath) {
		log.Printf("File already exists, skipping: %s", filePath)
		return
	}

	outFile, err := os.Create(filePath)
	if err != nil {
		log.Printf("Failed to create file %s: %v", filePath, err)
		return
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, resp.Body)
	if err != nil {
		log.Printf("Failed to save PDF to %s: %v", filePath, err)
		return
	}

	log.Printf("Downloaded to %s\n", filePath)
}

// fileExists checks if a file exists and is not a directory
func fileExists(filename string) bool {
	info, err := os.Stat(filename) // Get file info
	if err != nil {                // If stat fails (file doesn't exist)
		return false // Return false
	}
	return !info.IsDir() // Return true if it's a file, false if it's a directory
}

// Remove a file if it exists
func removeFileIfExists(filename string) {
	if fileExists(filename) { // Check if file exists
		err := os.Remove(filename) // Remove the file
		if err != nil {            // If removal fails
			log.Printf("Failed to remove file %s: %v", filename, err)
		} else {
			log.Printf("Removed file %s\n", filename) // Log successful removal
		}
	}
}

func main() {
	// Define input and output file paths here
	inputFile := "zsds3.zepinc.com.har"
	outputFile := "output.txt"

	// Open the input file for reading
	urlFromFile := extractURLsFromFileAndReturnSlice(inputFile)
	if urlFromFile == nil {
		log.Println("No URLs found in the input file")
		return
	}

	// Remove duplicates from the slice of URLs
	urlFromFile = removeDuplicatesFromSlice(urlFromFile)

	// Clean the URLs by validating and filtering them
	urlFromFile = cleanURLs(urlFromFile)

	outputDir := "zepPDF/" // Directory to save downloaded PDFs

	// Remove the output file if it exists
	removeFileIfExists(outputFile)

	// Write the URLs to the output file
	for _, url := range urlFromFile {
		// Download the PDF from the cleaned URL
		downloadPDF(url, outputDir)
		appendAndWriteToFile(outputFile, url)
	}
}
