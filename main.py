import os  # Import the os module for interacting with the operating system
import fitz  # Import PyMuPDF (fitz) for PDF handling (used for document validation)

# --- Utility Functions ---


def validate_pdf_file(
    file_path: str,
) -> bool:  # Define a function to validate a PDF file path, returning a boolean status
    """
    Attempts to open a PDF file using PyMuPDF (fitz) to check for corruption.
    """  # Start docstring explaining the function's purpose
    try:  # Start a try block to handle potential file reading errors
        # Using a context manager ensures the document object is properly closed.
        with fitz.open(
            file_path
        ) as document:  # Attempt to open the PDF file using fitz and assign the document object
            # Check if the PDF has at least one page
            if (
                document.page_count == 0
            ):  # Check the page count property of the document object
                print(
                    f"File '{file_path}': Invalid - Document has no pages."
                )  # Print an error message if the page count is zero
                return False  # Return False, indicating the PDF is invalid

            return True  # Return True if the document opened successfully and has pages (valid)
    except (
        RuntimeError
    ) as error_message:  # Catch a RuntimeError, which commonly indicates a corrupt PDF in PyMuPDF
        # This commonly catches errors like "cannot open document" for corrupt files
        print(
            f"File '{file_path}': Corrupt/Invalid - Error: {error_message}"
        )  # Print the specific error message
        return False  # Return False, indicating the PDF is invalid
    except FileNotFoundError:  # Catch the error if the file path does not exist
        print(
            f"File '{file_path}': Error - File not found."
        )  # Print a file not found error
        return False  # Return False for non-existent files
    except (
        Exception
    ) as unexpected_error:  # Catch any other unexpected exceptions during file opening
        # Catch any other unexpected error during opening
        print(
            f"File '{file_path}': Unexpected Error - {unexpected_error}"
        )  # Print the unexpected error details
        return False  # Return False, indicating an issue


def remove_system_file(
    system_path: str,
) -> bool:  # Define a function to safely remove a file
    """
    Safely removes a file from the system at the given path.
    """  # Docstring for the function
    try:  # Start a try block for file removal operations
        if os.path.exists(
            system_path
        ):  # Check if a file or directory exists at the given path
            os.remove(system_path)  # Use os.remove to delete the file
            return True  # Return True if removal was successful
        else:  # Execute if the file does not exist
            print(
                f"Removal failed: File not found at '{system_path}'"
            )  # Print a message indicating the file was not found
            return False  # Return False because nothing was removed
    except (
        PermissionError
    ):  # Catch the error if the program lacks permission to delete the file
        print(
            f"Removal failed: Permission denied for '{system_path}'"
        )  # Print a permission denied message
        return False  # Return False due to failure
    except Exception as error:  # Catch any other unexpected error during deletion
        print(
            f"Removal failed: An unexpected error occurred while deleting '{system_path}': {error}"
        )  # Print the unexpected deletion error
        return False  # Return False due to failure


def walk_directory_and_extract_given_file_extension(
    root_directory_path: str, file_extension: str
) -> list[str]:  # Define a function to recursively find files with a specific extension
    """
    Recursively walks a directory and returns a list of absolute paths
    for files ending with the specified extension (case-sensitive).
    """  # Docstring for the function
    matched_files: list[str] = (
        []
    )  # Initialize an empty list to store the absolute paths of matching files
    # Ensure extension starts with a dot for robust checking
    if not file_extension.startswith(
        "."
    ):  # Check if the extension string is missing the leading dot
        file_extension = "." + file_extension  # Prepend a dot if it's missing

    try:  # Start a try block for directory traversal
        # os.walk is used to traverse the directory tree
        for current_root, _, filenames in os.walk(
            root_directory_path
        ):  # Recursively traverse the directory tree
            # 'directories' is ignored (named with an unused placeholder in original code)
            for (
                file_name
            ) in (
                filenames
            ):  # Iterate over every file name found in the current directory
                if file_name.endswith(
                    file_extension
                ):  # Check if the file name ends with the specified extension
                    # os.path.join is used for creating system-independent paths
                    full_path = os.path.abspath(
                        os.path.join(current_root, file_name)
                    )  # Construct the full absolute path of the file
                    matched_files.append(
                        full_path
                    )  # Add the full path to the list of matched files
    except (
        FileNotFoundError
    ):  # Catch the error if the starting directory path is invalid
        print(
            f"Error: Directory not found at '{root_directory_path}'"
        )  # Print a directory not found error
    except (
        Exception
    ) as error:  # Catch any other unexpected errors during directory walk
        print(
            f"An error occurred while walking the directory: {error}"
        )  # Print the error details

    return matched_files  # Return the list of absolute paths


def get_filename_and_extension(
    full_path: str,
) -> str:  # Define a function to extract just the filename
    """
    Extracts only the filename and its extension from a full path.
    """  # Docstring for the function
    return os.path.basename(
        full_path
    )  # Use os.path.basename to return the final component of the path (the filename)


def check_upper_case_letter(
    content_string: str,
) -> bool:  # Define a function to check for uppercase letters
    """
    Checks if a string contains at least one uppercase letter.
    """  # Docstring for the function
    # Use a generator expression with 'any' for an efficient check
    return any(
        character.isupper() for character in content_string
    )  # Return True if any character in the string is uppercase


# --- Main Logic ---


def main():  # Define the main function to control the workflow
    """
    Main function to orchestrate the PDF validation and management workflow.
    """  # Docstring for the main function
    search_directory = (
        "./PDFs"  # Define the relative path to the directory to start searching
    )
    target_extension = ".pdf"  # Define the file extension to search for

    print(
        f"Starting PDF processing in directory: {search_directory}"
    )  # Inform the user about the starting directory
    print(
        f"Searching for files with extension: {target_extension}"
    )  # Inform the user about the target extension
    print("-" * 40)  # Print a separator line

    # Walk through the directory and extract .pdf files
    pdf_file_paths = walk_directory_and_extract_given_file_extension(
        search_directory, target_extension
    )  # Call the directory walking function to get a list of PDF paths

    if not pdf_file_paths:  # Check if the list of found files is empty
        print(
            f"No {target_extension} files found in '{search_directory}'. Exiting."
        )  # Print a message if no files were found
        return  # Exit the main function

    print(
        f"Found {len(pdf_file_paths)} files to process."
    )  # Report the total number of files found
    print("-" * 40)  # Print a separator line

    # Validate each PDF file
    for (
        file_path
    ) in pdf_file_paths:  # Iterate over the list of found absolute file paths
        file_name_with_ext = get_filename_and_extension(
            file_path
        )  # Extract the filename from the full path
        print(
            f"Processing: {file_name_with_ext}"
        )  # Print which file is currently being processed

        # 1. Check if the .PDF file is valid
        if not validate_pdf_file(
            file_path
        ):  # Call the validation function and check if it failed
            print(
                f"Invalid PDF detected: {file_path}. Attempting to delete file."
            )  # Log that an invalid PDF was found
            # Remove the invalid .pdf file.
            if remove_system_file(file_path):  # Attempt to delete the file
                print(
                    f"Successfully deleted invalid file: {file_name_with_ext}"
                )  # Log success if deletion worked
            continue  # Skip the rest of the loop block and move to the next file path

        # 2. Check for uppercase letters in the filename
        if check_upper_case_letter(
            file_name_with_ext
        ):  # Check if the filename contains any uppercase letters
            print(
                f"Warning: Uppercase letter(s) found in filename: {file_name_with_ext}"
            )  # Print a warning if uppercase letters are found
        else:
            print(
                f"Valid PDF and filename complies with case requirement."
            )  # Print success if the file is valid and filename is clean

        print("-" * 40)  # Print a separator line after processing each file


# Run the main function
if (
    __name__ == "__main__"
):  # Check if the script is being executed directly (standard entry point)
    main()  # Invoke the main function to start the PDF processing workflow
