package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"regexp"
	"slices"
	"sort"
)

// transformation config
type TransformationConfig struct {
	ColumnHeaders []string
	SortByColumn string
	SortAscending bool
	IncludeRegexByColumn map[string]string
	ExcludeRegexByColumn map[string]string
}
var transformation TransformationConfig = TransformationConfig{
	nil,
	"",
	true,
	make(map[string]string),
	make(map[string]string),
}

// presets
const presetsFilename = "presets"
var presetTransformations map[string]TransformationConfig = make(map[string]TransformationConfig)
var activePresetName string = ""

// output
var outputEntryIndices []int

func initializeTransformation() {

	// load presets from file
	if data, err := os.ReadFile(configDir + presetsFilename); err == nil {
		json.Unmarshal(data, &presetTransformations)
	}

	// if we have a preset, load that
	if *flags.preset != "" {
		if _, ok := presetTransformations[*flags.preset]; ok {
			usePreset(*flags.preset)
		} else {
			log.Fatalf("Could not use preset (%v) because it was not found.", *flags.preset)
		}

		return
	}

	// if we have a load path, load that
	if *flags.loadPath != "" {
		// load transformation file
		data, err := os.ReadFile(*flags.loadPath)
		if err != nil {
			log.Fatalf("Could not load transformation from file (%v): %v", *flags.loadPath, err)
		}

		// parse JSON
		if err := deserializeTransformation(data); err != nil {
			log.Fatalf("Could not parse loaded transformation %v", err)
		}

		return
	}

	// generate default transformation
	transformation.ColumnHeaders = make([]string, len(data.columnHeaders))
	copy(transformation.ColumnHeaders, data.columnHeaders)
}

func serializeTransformation() ([]byte, error) {
	out, error := json.MarshalIndent(transformation, "", "\t")
	if error != nil {
		return nil, error
	}
	return out, nil
}

func deserializeTransformation(data []byte) error {
	return json.Unmarshal(data, &transformation)
}

// PRESETS ====================================================================================================
func savePresetsToFile() {
	// marshal to json
	json, jsonError := json.MarshalIndent(&presetTransformations, "", "\t")
	if jsonError != nil {
		log.Fatalf("Could not marshal preset transformations to json: %v", jsonError)
	}

	// write json to file
	writeError := os.WriteFile(configDir + presetsFilename, json, 0644)
	if (writeError != nil) {
		log.Fatalf("Could not write presets to file: %v", writeError)
	}

	writeToMessageBuffer(fmt.Sprintf("Saved presets to %v", configDir + presetsFilename))
}

func usePreset(presetName string) {
	activePresetName = presetName
	transformation = deepCopyPreset(presetTransformations[presetName])
}

func deepCopyPreset(preset TransformationConfig) (out TransformationConfig) {
	copiedJson, err := json.Marshal(preset)
	if err != nil { panic(err) }

	json.Unmarshal(copiedJson, &out)
	return
}

// TRANSFORM LOGIC ===========================================================================================

func isColumnFake(header string) bool {
	return !slices.Contains(data.columnHeaders, header)
}

func getColumnFromData(header string) (column []string, fake bool) {
	column, ok := data.entriesByColumn[header]
	if ok {
		// real column, return
		return column, false
	} else {
		// make fake column
		column = make([]string, data.numEntries)
		for i := range column {
			column[i] = "NO DATA"
		}
		return column, true
	}
}

func getDataInColumn(header string, entry int) string {
	column, ok := data.entriesByColumn[header]
	if ok {
		return column[entry]
	} else {
		return "NO DATA"
	}
}


func filterInts(ss []int, test func(int) bool) (ret []int) {
    for _, e := range ss {
        if test(e) {
            ret = append(ret, e)
        }
    }
    return
}

func transformDataToOutput() {

	// generate default output (all of input)
	outputEntryIndices = make([]int, data.numEntries)
	for i := 0; i < len(outputEntryIndices); i++ { outputEntryIndices[i] = i }

	// filter by regex
	for _, columnHeader := range transformation.ColumnHeaders {
		if !slices.Contains(data.columnHeaders, columnHeader) { continue } // skip over if header not in data
		entries := data.entriesByColumn[columnHeader]

		// run include regex
		includeRegex, includeFound := transformation.IncludeRegexByColumn[columnHeader]
		if (includeFound) {
			compiledReg := regexp.MustCompile(includeRegex)
			outputEntryIndices = filterInts(outputEntryIndices, func(entryIndex int) bool {
				return compiledReg.MatchString(entries[entryIndex])
			})
		}

		// run exclude regex
		excludeRegex, excludeFound := transformation.ExcludeRegexByColumn[columnHeader]
		if (excludeFound) {
			compiledReg := regexp.MustCompile(excludeRegex)
			outputEntryIndices = filterInts(outputEntryIndices, func(entryIndex int) bool {
				return !compiledReg.MatchString(entries[entryIndex])
			})
		}
	}

	// sort
	columnToSortBy, found := data.entriesByColumn[transformation.SortByColumn]
	if (found) {
		sort.Slice(outputEntryIndices, func(i, j int) bool {
			valA := columnToSortBy[outputEntryIndices[i]] 
			valB := columnToSortBy[outputEntryIndices[j]]

			if transformation.SortAscending {
				return valA < valB
			} else {
				return valA > valB
			}
		})
	}
}
