import fitz  # PyMuPDF
import os # For file system operations

# Function to extract text from a PDF file using pymupdf
def extract_text_from_pdf_with_pymupdf(pdf_path):
    # Open the PDF file
    doc = fitz.open(pdf_path)
    # Extract text from all pages
    full_text = ""
    for page in doc:
        full_text += page.get_text()
    # Close the document
    doc.close()
    # Return the extracted text and page count
    return full_text

# This function saves a given string to a Markdown file with a specified name.
def save_to_md(content: str, file_path: str) -> None:
    with open(file_path, "w", encoding="utf-8") as md_file:
        md_file.write(content)

# Function to walk through a directory and extract files with a specific extension
def walkGivenDirectoryAndExtractCustomFileUsingFileExtension(system_path, extension):
    matched_files = []
    for root, _, files in os.walk(system_path):
        for file in files:
            if file.endswith(extension):
                full_path = os.path.abspath(os.path.join(root, file))
                matched_files.append(full_path)
    return matched_files

# Check if a file exists
def check_file_exists(system_path):
    return os.path.isfile(system_path)

def main():
    # Walk through the directory and extract .pdf files
    files = walkGivenDirectoryAndExtractCustomFileUsingFileExtension("./zepPDF", ".pdf")

    # Loop through each file and extract text
    for file_path in files:
        # Define the output Markdown file path
        md_file_path = os.path.splitext(file_path)[0] + ".md"
        
        # Check if the Markdown file already exists
        if check_file_exists(md_file_path):
            print(f"File {md_file_path} already exists. Skipping...")
            continue

        # Extract text from the PDF file
        content = extract_text_from_pdf_with_pymupdf(file_path)

        # Save the content to a Markdown file
        save_to_md(content, md_file_path)

        # Print a message indicating that the content has been saved
        print(f"Content saved to {md_file_path}")

main()
