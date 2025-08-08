package main

import (
	"fmt"
	"strings"
)

func printTable()  {
	// determine fluffing status
	shouldFluff := true
	if (*flags.noFluff) {
		shouldFluff = false
	}
	if (*flags.forceFluff) {
		shouldFluff = true
	}
	
	// determine column widths
	var widths []int = make([]int, len(transformation.ColumnHeaders))
	for c, header := range transformation.ColumnHeaders {
		// header width
		widths[c] = len(header)

		// entry widths
		column, _ := getColumnFromData(header)
		for _, rindex := range outputEntryIndices {
			valueLength := len(column[rindex])
			if (valueLength > widths[c]) {
				widths[c] = valueLength
			}
		}

		// add padding
		widths[c] += 2
	}

	// print headers
	for c, header := range transformation.ColumnHeaders {

		fake := isColumnFake(header)

		// apply padding
		header += strings.Repeat(" ", widths[c] - len(header))
		// colorize
		if (shouldFluff) { header = colorizeAnsiCell(header, true, c, fake) }

		fmt.Print(header)
	}
	fmt.Print("\n")

	// get output data (include fake columns)
	outColumns := make(map[string][]string)
	for _, header := range transformation.ColumnHeaders {
		column, _ := getColumnFromData(header)
		outColumns[header] = column
	}
	// print entries
	for _, entryIdx := range outputEntryIndices {
		for c, header := range transformation.ColumnHeaders {
			value := outColumns[header][entryIdx]

			// apply padding
			value += strings.Repeat(" ", widths[c] - len(value))
			// fluff
			if (shouldFluff) { value = colorizeAnsiCell(value, false, c, isColumnFake(header)) }

			fmt.Print(value)
		}
		fmt.Print("\n")
	}
}
