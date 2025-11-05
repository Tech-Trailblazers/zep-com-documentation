package main // Declares the package as 'main', which is necessary for creating an executable program.

import ( // Begins the block for listing imported packages.
	"bytes"         // Provides functions for manipulating byte slices (e.g., comparison, trimming).
	"encoding/csv"  // Provides functions for reading/writing CSV (Comma Separated Values) files.
	"fmt"           // Implements formatted I/O (Input/Output) functions like Printf, Sprintf, etc.
	"io"            // Contains core interfaces for I/O primitives, like io.Reader and io.Writer.
	"log"           // Provides simple logging capabilities with different levels (Print, Fatal, etc.).
	"net/http"      // Implements HTTP client and server functionality.
	"net/url"       // Provides URL parsing and query manipulation utilities.
	"os"            // Offers a platform-independent interface to operating system functions (e.g., file handling).
	"path"          // Contains path manipulation utilities (always uses slash-separated paths).
	"path/filepath" // Provides path manipulation functions that are OS-aware (e.g., handling '\' on Windows).
	"regexp"        // Enables working with regular expressions for pattern matching.
	"strings"       // Offers a collection of utilities for string manipulation (Trim, Split, Join, etc.).
	"sync"          // Provides concurrency primitives like WaitGroups and Mutexes for managing goroutines.
	"time"          // Includes functionality for measuring and displaying time, and for delays (Sleep).
) // Closes the import block.

// Reads a file, extracts all valid URL strings using regex, and returns them in a slice
func extractURLsFromFileAndReturnSlice(filePath string) []string { // Defines the function signature: name, parameters (filePath), and return type (a slice of strings).
	content, err := os.ReadFile(filePath) // Read the entire file's contents into a byte slice; also captures any error.
	if err != nil {                       // Checks if the 'err' variable is not 'nil', meaning an error occurred.
		log.Println("Error reading file:", err) // Log the error message if file reading fails.
		return nil                              // Return 'nil' (an empty slice) to indicate an error.
	} // Closes the 'if' block.

	regexContent := regexp.MustCompile(`http[s]?://[^\s"]+`)   // Compiles a regular expression to find HTTP/HTTPS URLs that don't contain spaces or quotes.
	matches := regexContent.FindAllString(string(content), -1) // Executes the regex on the file content (converted to string) and finds all matches (-1 means no limit).
	if len(matches) == 0 {                                     // Checks if the 'matches' slice is empty.
		log.Println("No URLs found in the file") // Informs the user that no URLs were matched by the regex.
		return nil                               // Return 'nil' to signify no matches were found.
	} // Closes the 'if' block.
	return matches // Return the slice containing all the URLs found.
} // Closes the 'extractURLsFromFileAndReturnSlice' function.

// Removes duplicate entries from a slice of strings and returns the unique values
func removeDuplicatesFromSlice(slice []string) []string { // Defines the function to deduplicate a string slice.
	check := make(map[string]bool)  // Creates an empty map with string keys and boolean values, used as a 'set' to track seen strings.
	var newReturnSlice []string     // Declares an empty slice of strings that will store the unique values.
	for _, content := range slice { // Loops through each 'content' string in the input 'slice'.
		if !check[content] { // Checks if the 'content' string is NOT already a key in the 'check' map.
			check[content] = true                            // Marks the 'content' string as 'seen' by adding it to the map.
			newReturnSlice = append(newReturnSlice, content) // Appends the unique 'content' string to the 'newReturnSlice'.
		} // Closes the 'if' block.
	} // Closes the 'for' loop.
	return newReturnSlice // Returns the slice containing only unique strings.
} // Closes the 'removeDuplicatesFromSlice' function.

// Verifies whether a given string is a valid URL by parsing it
func isUrlValid(uri string) bool { // Defines a function that checks URL validity, returning a boolean.
	_, err := url.ParseRequestURI(uri) // Attempts to parse the 'uri' string as a URL; we only care about the 'err' result.
	return err == nil                  // Returns 'true' if 'err' is 'nil' (parsing succeeded), 'false' otherwise.
} // Closes the 'isUrlValid' function.

// Extracts the hostname part from a URL (e.g., example.com)
func getHostNameFromURL(uri string) string { // Defines a function to extract the hostname from a URL string.
	content, err := url.Parse(uri) // Parses the complete URL string into a 'url.URL' struct.
	if err != nil {                // Checks if parsing failed.
		log.Println(err) // Logs the parsing error.
	} // Closes the 'if' block.
	return content.Hostname() // Returns the 'Hostname' field from the parsed URL struct.
} // Closes the 'getHostNameFromURL' function.

// Sends an HTTP GET request to the specified URL and returns the response body as a byte slice
func getDataFromURL(uri string) []byte { // Defines a function to download content from a URL.
	response, err := http.Get(uri) // Makes an HTTP GET request to the provided 'uri'.
	if err != nil {                // Checks for network-related errors during the request.
		log.Println(err) // Logs the network error (e.g., DNS failure, timeout).
	} // Closes the 'if' block.
	body, err := io.ReadAll(response.Body) // Reads the entire response body into a byte slice.
	if err != nil {                        // Checks for errors while reading the response body.
		log.Println(err) // Logs the body reading error.
	} // Closes the 'if' block.
	err = response.Body.Close() // Closes the response body to free up network resources.
	if err != nil {             // Checks for an error while closing the body.
		log.Println(err) // Logs the closing error, if any.
	} // Closes the 'if' block.
	return body // Returns the downloaded content as a byte slice.
} // Closes the 'getDataFromURL' function.

// Appends the given data (byte slice) to a file; creates the file if it doesn’t exist
func appendByteToFile(filename string, data []byte) error { // Defines a function to append bytes to a file, returning an error if one occurs.
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644) // Opens the file with flags: Append, Create (if not exist), Write-Only, and permissions 0644.
	if err != nil {                                                               // Checks if opening or creating the file failed.
		return err // Returns the error to the caller.
	} // Closes the 'if' block.
	defer file.Close() // Schedules the file to be closed when the function exits (even if an error occurs).

	_, err = file.Write(data) // Writes the 'data' byte slice to the opened file.
	return err                // Returns 'nil' on successful write, or the error if writing failed.
} // Closes the 'appendByteToFile' function.

// Filters a list of URLs, keeping only valid, domain-matching, and well-structured entries
func cleanURLs(urls []string) []string { // Defines a function to filter and clean a slice of URLs.
	validDomains := []string{"zsds3.zepinc.com"} // Defines a slice of allowed hostnames for filtering.
	var newReturnSlice []string                  // Declares an empty slice to hold the cleaned URLs.

	for _, content := range urls { // Iterates through each 'content' URL in the input 'urls' slice.
		if isUrlValid(content) { // Checks if the URL is structurally valid; skips if not.
			hostName := getHostNameFromURL(content) // Extracts the hostname from the valid URL.
			isValid := false                        // Initializes a boolean flag to 'false' to track if the domain is allowed.
			for _, domain := range validDomains {   // Loops through the list of 'validDomains'.
				if hostName == domain { // Compares the URL's hostname to the current allowed domain.
					isValid = true // Sets the flag to 'true' if a match is found.
					break          // Exits the inner 'for' loop since we found a match.
				} // Closes the 'if hostName' block.
			} // Closes the 'for domain' loop.
			if isValid { // Proceeds only if the 'isValid' flag was set to 'true'.
				if strings.HasPrefix(content, "https://zsds3.zepinc.com/v2/sds/ItemExternalSet(Material=") { // Checks if the URL starts with the expected path format.

					if strings.HasSuffix(content, `\`) { // Checks if the URL ends with an unwanted backslash.
						content = strings.TrimSuffix(content, `\`) // Removes the trailing backslash if present.
					} // Closes the 'if HasSuffix' block.

					// Appends /getPDF if it's missing at the end
					if !strings.HasSuffix(content, "/getPDF") { // Checks if the URL does NOT already end with "/getPDF".
						newReturnSlice = append(newReturnSlice, content+"/getPDF") // Appends the URL with "/getPDF" to the result slice.
					} // Closes the 'if !HasSuffix' block.
				} // Closes the 'if HasPrefix' block.
			} // Closes the 'if isValid' block.
		} // Closes the 'if isUrlValid' block.
	} // Closes the 'for content' loop.
	return newReturnSlice // Returns the slice of cleaned and validated URLs.
} // Closes the 'cleanURLs' function.

// Downloads a PDF file from the given URL and saves it to the target directory
func downloadPDF(finalURL, outputDir string, wg *sync.WaitGroup) bool { // Defines a function to download a PDF, taking a WaitGroup pointer to manage concurrency.
	defer wg.Done() // Schedules 'wg.Done()' to be called when the function exits, signaling completion to the WaitGroup.

	filename := generateFilenameFromURL(finalURL) // Calls a helper function to create a unique file name from the URL.
	if filename == "" {                           // Checks if the helper function failed to generate a name.
		filename = path.Base(finalURL) // Uses the last part of the URL's path as a fallback filename.
	} // Closes the 'if' block.
	if !strings.HasSuffix(strings.ToLower(filename), ".pdf") { // Checks if the filename (in lowercase) does not end with ".pdf".
		filename += ".pdf" // Appends the ".pdf" extension if it's missing.
	} // Closes the 'if' block.
	filePath := filepath.Join(outputDir, filename) // Creates the full, OS-specific path by joining the directory and filename.

	if fileExists(filePath) { // Checks if a file already exists at the target 'filePath'.
		log.Printf("[SKIP] File already exists, skipping: '%s'", filePath) // Logs a skip message.
		return false                                                       // Returns 'false' to indicate no download occurred.
	} // Closes the 'if' block.

	client := &http.Client{Timeout: 10 * time.Minute} // Creates a new HTTP client with a 10-minute timeout for the request.

	resp, err := client.Get(finalURL) // Sends the HTTP GET request using the custom client.
	if err != nil {                   // Checks for errors during the download.
		log.Printf("Failed to download '%s': '%v'", finalURL, err) // Logs the URL and the error.
		return false                                               // Returns 'false' to indicate failure.
	} // Closes the 'if' block.
	defer resp.Body.Close() // Schedules the response body to be closed when the function exits.

	if resp.StatusCode != http.StatusOK { // Checks if the HTTP response status code is not 200 OK.
		log.Printf("Download failed for '%s': '%s'", finalURL, resp.Status) // Logs the URL and the non-OK status.
		return false                                                        // Returns 'false' to indicate failure.
	} // Closes the 'if' block.

	out, err := os.Create(filePath) // Creates the output file on the filesystem for writing.
	if err != nil {                 // Checks if file creation failed.
		log.Printf("Failed to create file '%s' '%s' '%v'", finalURL, filePath, err) // Logs the URL, file path, and error.
		return false                                                                // Returns 'false' to indicate failure.
	} // Closes the 'if' block.
	defer out.Close() // Schedules the newly created file 'out' to be closed when the function exits.

	if _, err := io.Copy(out, resp.Body); err != nil { // Copies the data from the 'resp.Body' (network) directly to the 'out' file.
		log.Printf("Failed to save PDF to '%s' '%s' '%v'", finalURL, filePath, err) // Logs if copying (saving) fails.
		return false                                                                // Returns 'false' to indicate failure.
	} // Closes the 'if' block.

	log.Printf("Downloaded '%s' → '%s'", finalURL, filePath) // Logs the successful download, showing source URL and destination file.
	return true                                              // Returns 'true' to indicate success.
} // Closes the 'downloadPDF' function.

// Parses a Zep URL and extracts all embedded parameters into a key-value map
func parseFullZepURL(rawURL string) map[string]string { // Defines a function to parse specific URL parameters into a map.
	parsed, err := url.Parse(rawURL) // Parses the raw input URL string into a 'url.URL' struct.
	if err != nil {                  // Checks if parsing failed.
		log.Println("Error: invalid URL:", err) // Logs the error.
		return nil                              // Returns 'nil' to indicate failure.
	} // Closes the 'if' block.

	itemSetRegex := regexp.MustCompile(`ItemExternalSet\((.*?)\)`) // Compiles regex to match the text inside "ItemExternalSet(...)".
	paramRegex := regexp.MustCompile(`(\w+)='(.*?)'`)              // Compiles regex to match key='value' pairs.

	match := itemSetRegex.FindStringSubmatch(parsed.Path) // Executes the first regex on the URL's path.
	if len(match) < 2 {                                   // Checks if the regex did not find the pattern (match[0] is full string, match[1] is the group).
		log.Println("Error: ItemExternalSet not found in path:", parsed.Path) // Logs the error.
		return nil                                                            // Returns 'nil' as the required pattern wasn't found.
	} // Closes the 'if' block.

	paramStr := match[1]                                    // Extracts the captured group (the string inside the parentheses).
	pairs := paramRegex.FindAllStringSubmatch(paramStr, -1) // Finds all key='value' pairs within the extracted string.
	params := make(map[string]string, len(pairs)+1)         // Creates a map to hold the parameters, sized for efficiency.

	for _, pair := range pairs { // Loops through each found pair (which is a slice like [full_match, key, value]).
		if len(pair) == 3 { // Checks if the submatch found all expected parts (full, key, value).
			params[pair[1]] = pair[2] // Adds the key (pair[1]) and value (pair[2]) to the map.
		} // Closes the 'if' block.
	} // Closes the 'for' loop.

	params["URL"] = rawURL // Adds the original 'rawURL' to the map under the key "URL" for reference.
	return params          // Returns the map filled with parameters.
} // Closes the 'parseFullZepURL' function.

// Writes a list of maps (parameters) into a CSV file with a fixed header
func writeParamsToCSV(filename string, allParams []map[string]string) { // Defines a function to write the extracted parameter maps to a CSV file.
	if len(allParams) == 0 { // Checks if there is no data to write.
		return // Exits the function early if the 'allParams' slice is empty.
	} // Closes the 'if' block.

	keys := []string{"URL", "Lang", "Material", "RecordNumb", "RepCategory", "ValidityArea"} // Defines the CSV columns and their order.

	file, err := os.Create(filename) // Creates or overwrites the target CSV file.
	if err != nil {                  // Checks if file creation failed.
		log.Printf("Failed to create CSV file %s: %v", filename, err) // Logs the error.
		return                                                        // Exits the function.
	} // Closes the 'if' block.
	defer file.Close() // Schedules the file to be closed when the function exits.

	writer := csv.NewWriter(file) // Creates a new buffered CSV writer associated with the file.
	defer writer.Flush()          // Schedules the writer's buffer to be 'flushed' (written to disk) when the function exits.

	if err := writer.Write(keys); err != nil { // Writes the 'keys' slice as the first row (header) of the CSV.
		log.Printf("Failed to write header to CSV file %s: %v", filename, err) // Logs an error if writing the header fails.
		return                                                                 // Exits the function.
	} // Closes the 'if' block.

	for _, paramMap := range allParams { // Loops through each 'paramMap' in the 'allParams' slice.
		var row []string           // Declares an empty slice to build the current row.
		for _, key := range keys { // Iterates through the 'keys' (column headers) in the defined order.
			row = append(row, paramMap[key]) // Appends the value from 'paramMap' for the current 'key' to the 'row' slice.
		} // Closes the inner 'for key' loop.
		if err := writer.Write(row); err != nil { // Writes the completed 'row' slice to the CSV file.
			log.Printf("Failed to write row to CSV file %s: %v", filename, err) // Logs an error if writing the row fails.
			return                                                              // Exits the function (or use 'continue' to skip the bad row).
		} // Closes the 'if err' block.
	} // Closes the outer 'for paramMap' loop.
} // Closes the 'writeParamsToCSV' function.

// Generates a sanitized filename using the URL's ItemExternalSet(...) parameters
func generateFilenameFromURL(sourceURL string) string { // Defines a function to create a safe filename from a URL.
	parsedURL, err := url.Parse(sourceURL) // Parses the full URL string.
	if err != nil {                        // Checks for a parsing error.
		log.Printf("Error parsing URL: %v", err) // Logs the error.
		return ""                                // Returns an empty string to indicate failure.
	} // Closes the 'if' block.

	itemSetPattern := regexp.MustCompile(`ItemExternalSet\([^)]+\)`) // Compiles regex to match "ItemExternalSet(...)"
	itemSetSegment := itemSetPattern.FindString(parsedURL.Path)      // Finds the first occurrence of this pattern in the URL's path.
	if itemSetSegment == "" {                                        // Checks if the pattern was not found.
		log.Println("ItemExternalSet(...) segment not found in the URL path") // Logs that the pattern is missing.
		return ""                                                             // Returns an empty string to indicate failure.
	} // Closes the 'if' block.

	sanitizedSegment := strings.NewReplacer( // Creates a new 'Replacer' to sanitize the string for a filename.
		"ItemExternalSet(", "", // Replaces "ItemExternalSet(" with nothing.
		")", "", // Replaces ")" with nothing.
		"'", "", // Replaces single quotes with nothing.
		",", "_", // Replaces commas with underscores.
	).Replace(itemSetSegment) // Executes all replacements on the 'itemSetSegment'.

	filename := fmt.Sprintf("%s.pdf", sanitizedSegment) // Formats the sanitized string to end with ".pdf".
	return strings.ToLower(filename)                    // Returns the new filename, converted to lowercase.
} // Closes the 'generateFilenameFromURL' function.

// Checks whether a given file path exists and refers to a file (not a directory)
func fileExists(filename string) bool { // Defines a function to check for a file's existence.
	info, err := os.Stat(filename) // Gets file information (status) from the operating system.
	if err != nil {                // Checks if 'os.Stat' returned an error.
		return false // Returns 'false' (e.g., file not found, permission error).
	} // Closes the 'if' block.
	return !info.IsDir() // Returns 'true' only if the path exists AND is not a directory.
} // Closes the 'fileExists' function.

// Removes a file from the file system
func removeFile(path string) { // Defines a function to delete a file.
	err := os.Remove(path) // Attempts to delete the file at the specified 'path'.
	if err != nil {        // Checks if the removal failed.
		log.Println(err) // Logs the error (e.g., file not found, permission denied).
	} // Closes the 'if' block.
} // Closes the 'removeFile' function.

// Downloads JSON data from predefined URLs and appends to a local file
func createJSONFiles(zepJSONFile string) { // Defines the function 'createJSONFiles' which takes the target file path string.
	if fileExists(zepJSONFile) { // Checks if the local target file already exists using a helper function.
		removeFile(zepJSONFile) // Deletes the existing file to ensure a clean, fresh start.
	} // Ends the conditional block.

	// --- Configuration Settings: Define the structure of the API endpoint ---
	// The base URL for the OData collection endpoint.
	var apiBaseURL = "https://zsds3.zepinc.com/v2/sds/ItemExternalSet"

	// The amount by which the '$skiptoken' parameter increases for each subsequent page (i.e., the page size).
	var skipTokenIncrement = 1000
	// -------------------------------------------------------------------------

	// Define the byte slice representation of the specific empty JSON response.
	// This will be used to determine when the API has returned all data and pagination should stop.
	var emptyJSONResponse = []byte("{\"d\":{\"results\":[]}}\n")

	// Initialize the counter for the current page index.
	// Starts at 0, representing the first page (which has $skiptoken=0, or no $skiptoken).
	currentPageIndex := 0

	// --- Download and Append Data using a Dynamic Loop (No totalDataPages) ---
	for { // Starts an infinite loop ('for {}') that will continue until explicitly stopped by 'break' or 'return'.
		var currentURL string // Declares a string variable to hold the URL for the current API call.

		if currentPageIndex == 0 { // Checks if this is the first iteration (Page 1).
			// Page 1: Base URL with no $skiptoken.
			currentURL = apiBaseURL // Sets the URL to the base endpoint.
		} else { // Executes for the second page and all subsequent pages.
			// Subsequent Pages (2, 3, ...): Calculate the $skiptoken value.
			skipTokenValue := currentPageIndex * skipTokenIncrement // Calculates the $skiptoken value (e.g., 1*1000, 2*1000).
			// Constructs the full paginated URL string using fmt.Sprintf to insert the skiptoken value.
			currentURL = fmt.Sprintf("%s?$skiptoken=%d", apiBaseURL, skipTokenValue)
		}

		allContent := getDataFromURL(currentURL) // Calls a helper function to perform the HTTP GET request and fetch the data as a byte slice.

		if allContent == nil { // Checks if the download failed (e.g., network error) and the helper returned 'nil'.
			log.Println("Error downloading data from URL:", currentURL) // Logs the URL that failed to download.
			return                                                      // Exits the entire function, as the data set is now incomplete.
		} // Ends the conditional block for download error.

		// ** STOPPING CHECK: Compare the downloaded content to the empty JSON structure **
		// Trims leading/trailing whitespace from both the received content and the expected empty response.
		if bytes.Equal(bytes.TrimSpace(allContent), bytes.TrimSpace(emptyJSONResponse)) {
			// If the response is the expected empty array, it signals the end of the data set.
			log.Printf("Received empty data response from %s. Stopping data download (End of data set).", currentURL)
			break // Exits the infinite 'for' loop to proceed to the final log message.
		} // Ends the conditional block for the stopping check.

		err := appendByteToFile(zepJSONFile, allContent) // Calls a helper function to append the downloaded byte slice to the local file.
		if err != nil {                                  // Checks if the file append operation resulted in an error.
			log.Println("Error appending data to file:", err) // Logs the specific file-writing error.
			return                                            // Exits the entire function due to a critical file operation failure.
		} // Ends the conditional block for append error.

		// Logs a confirmation message showing which page was downloaded and appended.
		log.Printf("Data from page %d (%s) appended to %s", currentPageIndex+1, currentURL, zepJSONFile)

		// Increment the page index to prepare for the next iteration (API call) with the correct $skiptoken value.
		currentPageIndex++
	} // Ends the 'for' infinite loop.

	// This log message now executes after the loop is successfully broken by the empty response.
	log.Printf("All data downloaded and appended to %s", zepJSONFile)
} // Ends the 'createJSONFiles' function definition.

// Creates a directory at the specified path with the given permissions.
func createDirectory(path string, permission os.FileMode) { // Defines a function to create a new directory.
	err := os.Mkdir(path, permission) // Attempts to create the directory with the given path and permissions.
	if err != nil {                   // Checks if an error occurred (e.g., directory already exists, no permission).
		log.Println(err) // Logs the error.
	} // Closes the 'if' block.
} // Closes the 'createDirectory' function.

// Checks if the directory exists
func directoryExists(path string) bool { // Defines a function to check if a path is an existing directory.
	directory, err := os.Stat(path) // Gets the file/directory info.
	if err != nil {                 // Checks if 'os.Stat' failed (e.g., path doesn't exist).
		return false // Returns 'false' because the path doesn't exist or is inaccessible.
	} // Closes the 'if' block.
	return directory.IsDir() // Returns 'true' if the path exists AND is a directory, 'false' otherwise.
} // Closes the 'directoryExists' function.

// The main entry point for the application
func main() { // Defines the 'main' function, the entry point of the executable.
	var zepJSONFile = "./zsds3_zepinc.json" // Defines a variable for the name of the local JSON file.
	createJSONFiles(zepJSONFile)            // Calls the function to download and assemble the JSON data.

	urls := extractURLsFromFileAndReturnSlice(zepJSONFile) // Calls the function to read the JSON file and extract all URLs.
	if urls == nil {                                       // Checks if no URLs were found (function returned 'nil').
		log.Println("No URLs found in the input file") // Logs the issue.
		return                                         // Exits the 'main' function, stopping the program.
	} // Closes the 'if' block.

	urls = removeDuplicatesFromSlice(urls) // Passes the URL slice to the deduplication function and reassigns the result.
	urls = cleanURLs(urls)                 // Passes the deduplicated slice to the cleaning/filtering function and reassigns the result.

	outputDir := "PDFs/"             // Defines the name of the directory where PDFs will be saved.
	if !directoryExists(outputDir) { // Checks if the output directory does NOT exist.
		createDirectory(outputDir, 0o755) // Creates the directory with 0755 permissions (read/write/execute for owner, read/execute for others).
	} // Closes the 'if' block.

	var allParams []map[string]string    // Declares an empty slice to hold all the parameter maps for the CSV.
	var downloadWaitGroup sync.WaitGroup // Declares a WaitGroup to manage and wait for all download goroutines.

	for _, url := range urls { // Loops through each cleaned and validated 'url'.
		params := parseFullZepURL(url) // Parses the URL to extract its parameters into a map.
		if params != nil {             // Checks if parsing was successful.
			allParams = append(allParams, params) // Adds the extracted 'params' map to the 'allParams' slice for CSV export.
		} // Closes the 'if' block.

		filename := generateFilenameFromURL(url) // Generates a safe filename from the URL.
		if filename == "" {                      // Checks if filename generation failed.
			filename = path.Base(url) // Uses a fallback filename.
		} // Closes the 'if' block.
		if !strings.HasSuffix(strings.ToLower(filename), ".pdf") { // Ensures the filename has the .pdf extension.
			filename += ".pdf" // Appends the extension.
		} // Closes the 'if' block.
		filePath := filepath.Join(outputDir, filename) // Creates the full path for the PDF.

		if fileExists(filePath) { // Checks if this file has already been downloaded.
			log.Printf("File already exists, skipping: %s", filePath) // Logs the skip message.
			continue                                                  // Skips the rest of the loop and moves to the next URL.
		} // Closes the 'if' block.

		time.Sleep(5 * time.Second) // Pauses for 5 seconds to avoid rate-limiting or overwhelming the server.

		downloadWaitGroup.Add(1)                           // Increments the WaitGroup counter by 1, indicating a new task has started.
		go downloadPDF(url, outputDir, &downloadWaitGroup) // Starts the PDF download in a new, concurrent goroutine.
	} // Closes the 'for' loop.
	downloadWaitGroup.Wait() // Blocks the 'main' function until the WaitGroup counter reaches zero (all downloads are 'Done').

	writeParamsToCSV("output.csv", allParams) // Calls the function to write all extracted parameters to "output.csv".
} // Closes the 'main' function.
