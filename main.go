package main

import (
	"bufio"
	"devopsdb/connectors"
	"devopsdb/engine"
	"devopsdb/inputs"
	"fmt"
	"math"
	"os"
	"strings"
)

// This is a test bed for the real implementation
// Unlike the rest of the code, there are no tests around this area and it will be removed
// in favour of a proper command-line interface
func main() {

	// Init the whole thing
	engine := engine.New()
	engine.AddConnector(
		"devops",
		connectors.CreateDevopsClient(
			"",
			"",
		),
	)

	fmt.Println("Enter your query:")

	reader := bufio.NewReader(os.Stdin)
	// ReadString will block until the delimiter is entered
	input, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Error while reading query.", err)
		return
	}

	// remove the delimeter from the string
	input = strings.TrimSuffix(input, "\n")

	query, _ := inputs.SqlToQuery(input)

	result := engine.Execute(query)

	if len(result.Results) == 0 {
		fmt.Print("No results")
		os.Exit(0)
	}

	fmt.Println()
	fmt.Printf("%v results:\n", len(result.Results))
	fmt.Println()

	// Get column sizes
	columnSizes := make(map[string]int)

	for _, item := range result.Results {
		for column, _ := range item {
			if columnSizes[column] == 0 || columnSizes[column] < len(item[column]) {
				columnSizes[column] = len(item[column])
			}
		}
	}

	totalLength := 0
	for _, size := range columnSizes {
		totalLength += size
	}

	// Print column headers
	fmt.Print("| ")
	for _, column := range result.Columns {
		fmt.Print(StrPad(column, columnSizes[column], " ", "RIGHT"))
		fmt.Print(" | ")
	}

	fmt.Println()
	// Underline
	fmt.Print("|-")
	for _, column := range result.Columns {
		fmt.Print(StrPad("-", columnSizes[column], "-", "RIGHT"))
		fmt.Print("-| ")
	}
	fmt.Println()

	// Values
	for _, item := range result.Results {
		fmt.Print("| ")

		for _, column := range result.Columns {
			val := item[column]
			fmt.Print(StrPad(val, columnSizes[column], " ", "RIGHT"))
			fmt.Print(" | ")
		}
		fmt.Println()
	}

	fmt.Println()
}

// StrPad returns the input string padded on the left, right or both sides using padType to the specified padding length padLength.
//
// Example:
// input := "Codes";
// StrPad(input, 10, " ", "RIGHT")        // produces "Codes     "
// StrPad(input, 10, "-=", "LEFT")        // produces "=-=-=Codes"
// StrPad(input, 10, "_", "BOTH")         // produces "__Codes___"
// StrPad(input, 6, "___", "RIGHT")       // produces "Codes_"
// StrPad(input, 3, "*", "RIGHT")         // produces "Codes"
func StrPad(input string, padLength int, padString string, padType string) string {
	var output string

	inputLength := len(input)
	padStringLength := len(padString)

	if inputLength >= padLength {
		return input
	}

	repeat := math.Ceil(float64(1) + (float64(padLength-padStringLength))/float64(padStringLength))

	switch padType {
	case "RIGHT":
		output = input + strings.Repeat(padString, int(repeat))
		output = output[:padLength]
	case "LEFT":
		output = strings.Repeat(padString, int(repeat)) + input
		output = output[len(output)-padLength:]
	case "BOTH":
		length := (float64(padLength - inputLength)) / float64(2)
		repeat = math.Ceil(length / float64(padStringLength))
		output = strings.Repeat(padString, int(repeat))[:int(math.Floor(float64(length)))] + input + strings.Repeat(padString, int(repeat))[:int(math.Ceil(float64(length)))]
	}

	return output
}
