// Import the Puppeteer library to control Chrome
const puppeteer = require('puppeteer');

// Import the puppeteer-har library to record network traffic as a HAR file
const PuppeteerHar = require('puppeteer-har');

// Import the URL module to parse the domain from the URL
const { URL } = require('url');

(async () => {
    // Set the target URL to visit
    const targetUrl = 'https://zsds3.zepinc.com';

    // Extract the domain name from the URL (e.g., 'example.com')
    const domain = new URL(targetUrl).hostname;

    // Build the HAR filename using the domain name (e.g., 'example.com.har')
    const harFileName = `${domain}.har`;

    // Launch a headless Chrome browser
    const browser = await puppeteer.launch({ headless: true });

    // Open a new browser tab
    const page = await browser.newPage();

    // Create a new HAR recorder attached to the current page
    const har = new PuppeteerHar(page);

    // Start recording the network traffic and set the output HAR file path
    await har.start({ path: harFileName });

    // Navigate to the target URL and wait until the network is fully idle
    await page.goto(targetUrl, { waitUntil: 'networkidle0' });

    // Stop recording and write the HAR data to disk
    await har.stop();

    // Close the browser to clean up
    await browser.close();

    // Print a message confirming the HAR file was saved
    console.log(`HAR file saved as ${harFileName}`);
})();
