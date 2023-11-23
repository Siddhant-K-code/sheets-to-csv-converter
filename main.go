package main

import (
	"encoding/csv"
	"errors"
	"fmt"
	"os"

	"github.com/jessevdk/go-flags"
	"github.com/xuri/excelize/v2"
)

// Options struct holds command line options
type Options struct {
	ListSheets bool   `short:"l" long:"list-sheets" description:"List sheets"`
	Sheet      string `short:"s" long:"sheet" description:"Sheet to convert"`
	InputFile  string `short:"i" long:"input" description:"Input XLSX file (default: stdin)"`
	OutputFile string `short:"o" long:"output" description:"Output CSV file (default: stdout)"`
}

var options Options

func main() {
	// Parse command line flags
	if _, err := flags.Parse(&options); err != nil {
		handleCommandLineErrors(err)
	}

	// Run the main CLI logic
	if err := executeCli(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// handleCommandLineErrors processes errors related to command line flag parsing
func handleCommandLineErrors(err error) {
	if flagsErr, ok := err.(flags.ErrorType); ok && flagsErr == flags.ErrHelp {
		os.Exit(0)
	}
	os.Exit(1)
}

// executeCli is the main logic for the CLI application
func executeCli() error {
	// Open the Excel file for processing
	excelFile, err := openExcelFileForProcessing(options.InputFile)
	if err != nil {
		return err
	}
	defer safelyCloseFile(excelFile)

	// List sheets if the option is set, otherwise convert to CSV
	if options.ListSheets {
		printSheetNames(excelFile)
		return nil
	}

	return convertExcelToCSV(excelFile)
}

// openExcelFileForProcessing opens an Excel file for reading, defaulting to stdin if no file is provided
func openExcelFileForProcessing(path string) (*excelize.File, error) {
	if path == "" {
		return excelize.OpenReader(os.Stdin)
	}
	return excelize.OpenFile(path)
}

// printSheetNames lists all sheet names in the Excel file
func printSheetNames(f *excelize.File) {
	for _, sheet := range f.GetSheetList() {
		fmt.Println(sheet)
	}
}

// convertExcelToCSV converts the specified Excel sheet to a CSV file
func convertExcelToCSV(f *excelize.File) error {
	// Select the appropriate sheet to convert
	sheetName, err := determineSheetName(f, options.Sheet)
	if err != nil {
		return err
	}

	// Retrieve the rows from the sheet
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return handleSheetReadError(err, sheetName)
	}

	return outputRowsToCSV(rows)
}

// determineSheetName decides which sheet to use for conversion
func determineSheetName(f *excelize.File, sheetName string) (string, error) {
	if sheetName != "" {
		return sheetName, nil
	}

	sheetNames := f.GetSheetList()
	if len(sheetNames) == 0 {
		return "", errors.New("no sheets found in file")
	}

	firstSheet := sheetNames[0]
	fmt.Fprintf(os.Stderr, "no sheet specified, using '%s'\n", firstSheet)
	return firstSheet, nil
}

// handleSheetReadError processes errors related to reading a specific Excel sheet
func handleSheetReadError(err error, sheetName string) error {
	if _, ok := err.(excelize.ErrSheetNotExist); ok {
		return fmt.Errorf("sheet '%s' does not exist", sheetName)
	}
	return err
}

// outputRowsToCSV writes the provided rows to a CSV file
func outputRowsToCSV(rows [][]string) error {
	// Determine the output file for the CSV data
	outputFile, err := determineOutputFile(options.OutputFile)
	if err != nil {
		return err
	}
	defer safelyCloseFile(outputFile)

	// Create a CSV writer and write the rows
	csvWriter := csv.NewWriter(outputFile)
	defer csvWriter.Flush()

	for _, row := range rows {
		if err := csvWriter.Write(row); err != nil {
			return err
		}
	}

	return nil
}

// determineOutputFile opens or creates a file for writing, defaulting to stdout if no file is specified
func determineOutputFile(filePath string) (*os.File, error) {
	if filePath == "" {
		return os.Stdout, nil
	}
	return os.Create(filePath)
}

// safelyCloseFile safely closes a file resource, handling any errors
func safelyCloseFile(closer interface{ Close() error }) {
	if err := closer.Close(); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}
