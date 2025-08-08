package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
)

// flags
var flags = struct {
	command *string
	loadPath *string
	preset *string
	parseMode *string
	fatTable *bool
	stdout *bool
	noFluff *bool
	forceFluff *bool
	forceTui *bool
}{
	flag.String("command", "", "Command used to fetch table."),
	flag.String("load", "", "Path to load transformation from."),
	flag.String("p", "", "Name of preset to load"),
	flag.String("parseMode", "positional", "Table parsing mode. 'whitespace' or 'positional' are accepted."),
	flag.Bool("fatTable", false, "Table display mode. Enable to turn on table borders."),
	flag.Bool("stdout", false, "Print output instead of displaying TUI. This will effectively do nothing if you don't load a transformation as well."),
	flag.Bool("noFluff", false, "Enable to disable fluff (colors and symbols) in the stdout output."),
	flag.Bool("forceFluff", false, "Enable to force fluff. Bypasses automatic disablement of fluff when piping to other programs."),
	flag.Bool("forceTui", false, "Enable to enter the tui even when piping to other programs."),
}

// input data
var data = struct {
	entriesByColumn map[string][]string
	columnHeaders []string
	numEntries int
	numHeaderRows int
}{
	nil,
	nil,
	0,
	1,
}

func main() {

	// parse flags
	flag.Parse()

	// validate flags
	if *flags.parseMode != "positional" && *flags.parseMode != "whitespace" {
		fmt.Println("Bad parse mode")
		os.Exit(1)
	}
	// detect std out
    if fi, _ := os.Stdout.Stat(); (fi.Mode() & os.ModeCharDevice) == 0 {
		*flags.stdout = true
		*flags.noFluff = true
	}
	if *flags.forceTui {
		*flags.stdout = false
	}

	// initialize config
	initializeConfig()

	// get input from stdin or running command
	var inputText string
    if fi, _ := os.Stdin.Stat(); (fi.Mode() & os.ModeCharDevice) == 0 {
		// read input from std in
		bytes, err := io.ReadAll(os.Stdin)
		if err != nil {
			fmt.Println("Failed to read stdin")
			os.Exit(1)
		}
		inputText = string(bytes)
	} else if *flags.command != "" {
		// read input from command
		inputText = getCommandOutput(*flags.command)
	} else {
		log.Fatal("Could not find an input source. Please use the -command flag or stdin.")
	}

	// parse text into input entries
	parseInput(inputText)

	// load or generate default transformation
	initializeTransformation()

	// generate output
	transformDataToOutput()

	// run TUI
	if !*flags.stdout {
		setupTui()
	} else {
		printTable()
	}
}

func parseInput(inputString string) {
	data.entriesByColumn = make(map[string][]string) 

	lines := strings.Split(inputString, "\n")
	data.columnHeaders = strings.Fields(lines[0])

	// calculate header start indices (if positional)
	var headerStartIndices []int = make([]int, len(data.columnHeaders))
	if *flags.parseMode == "positional" {
		lastHeaderEnd := 0
		for i, header := range data.columnHeaders {
			if i == 0 {
				headerStartIndices[i] = 0
				lastHeaderEnd = len(header)
				continue
			}

			// search for next header index after last header
			searchString := lines[0][lastHeaderEnd:]
			headerStartIndices[i] = strings.Index(searchString, header) + lastHeaderEnd
			lastHeaderEnd = headerStartIndices[i] + len(header)
		}
	}

	// iterate over each entry in the table
	data.numEntries = 0
	for i := 1; i < len(lines); i++ {
		// skip over empty line (the last one) (also skips the entry count from increasing)
		if (len(lines[i]) == 0) { continue }

		// parse column entries (depending on mode)
		if *flags.parseMode == "positional" {
			// positional mode
			// go through each position
			for vindex, position := range headerStartIndices {
				var value string
				if (vindex < len(headerStartIndices) - 1) {
					// normally we can go from the current position to the next
					value = lines[i][position:headerStartIndices[vindex + 1]]
				} else {
					// if we are on the last header index, go until the end
					value = lines[i][position:]
				}
				data.entriesByColumn[data.columnHeaders[vindex]] = append(data.entriesByColumn[data.columnHeaders[vindex]], strings.TrimSpace(value))
			}

		} else {
			values := strings.Fields(lines[i])
			// whitespace mode
			// iterate through each value in the row and add it to the list
			for vindex, value := range values {
				data.entriesByColumn[data.columnHeaders[vindex]] = append(data.entriesByColumn[data.columnHeaders[vindex]], value)
			}
		}

		// increment entry count
		data.numEntries++
	}
}

func getCommandOutput(command string) string {
	// get command output
	cmd := exec.Command("sh", "-c", command)
	stdout, err := cmd.Output()
	if err != nil {
		fmt.Println(err.Error())
		return ""
	}

	return string(stdout)
}
