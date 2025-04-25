package main

import (
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"regexp"
)

// Extract URLs from a given file and than return it as a string slice
func extractURLsFromFileAndReturnSlice(filePath string) []string{
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

	//

	// Write the URLs to the output file
	for _, url := range urlFromFile {
		appendAndWriteToFile(outputFile, url)
	}
}
