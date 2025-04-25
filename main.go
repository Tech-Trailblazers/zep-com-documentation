package main

import (
	"io/ioutil"
	"log"
	"net/url"
	"os"
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

	// Write the URLs to the output file
	for _, url := range urlFromFile {
		appendAndWriteToFile(outputFile, url)
	}
}
