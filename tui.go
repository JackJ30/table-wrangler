package main

import (
	"fmt"
	"log"
	"os"
	"slices"

	"golang.design/x/clipboard"

	// external
	"github.com/gdamore/tcell/v2"
	"github.com/JackJ30/jack-tview"
)

// live data
var controlPanelVisible = true

// global tui objects
var app *tview.Application
var table *tview.Table
var pages *tview.Pages
var messageBuffer *tview.TextView
var instructionsText *tview.TextView
var infoText *tview.TextView

func setupTui()  {
	// initialize clipboard lib
	err := clipboard.Init()
	if err != nil {
		fmt.Println("Using the clipboard is not supported")
	}
	
	// initialize tview
	app = tview.NewApplication()
	app.EnableMouse(true)

	// main flex view
	flex := tview.NewFlex()

	// left side (messageBuffer and pages)
	leftFlex := tview.NewFlex().SetDirection(tview.FlexRow)
	flex.AddItem(leftFlex, 0, 5, true)
	pages = tview.NewPages()
	leftFlex.AddItem(pages, 0, 1, true)
	messageBuffer = tview.NewTextView().SetText("")
	messageBuffer.SetBackgroundColor(tcell.ColorDimGrey)
	leftFlex.AddItem(messageBuffer, 1, 0, false)

	// main table view
	createTable()
	pages.AddPage("main", table, true, true)

	// Control panel
	controlPanel := tview.NewFlex().SetDirection(0)
	controlPanel.SetBorder(true).SetTitle("Jack's Table Wrangler")
	flex.AddItem(controlPanel, 0, 1, false)
	setControlPanelVisibility := func()  {
		if controlPanelVisible {
			flex.ResizeItem(controlPanel, 0, 1)
		} else {
			flex.ResizeItem(controlPanel, 0, 0)
		}
	}
	setControlPanelVisibility()

	// Control Panel Instructions text
	instructionsText = tview.NewTextView().SetDynamicColors(true).SetWrap(true)
	controlPanel.AddItem(instructionsText, 0, 1, false)

	// Control Panel Info:
	infoText = tview.NewTextView().SetDynamicColors(true).SetWrap(true)
	controlPanel.AddItem(infoText, 0, 1, false)
	updateInfoText()

	// app level input handling
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyCtrlH:
			controlPanelVisible = !controlPanelVisible
			setControlPanelVisibility()
			return nil
		case tcell.KeyCtrlC:
			exitTui()
		}

		return event
	})

	// modes
	appendMode(ColumnMode{})

	// run tview
	if err := app.SetRoot(flex, true).Run(); err != nil {
		panic(err)
	}
}

// UTILITIES ================================================================================

func updateInfoText() {
	infoText.SetText(fmt.Sprintf("[orange::b]Info[w::-]\nNum entries (after filter): %v", len(outputEntryIndices)))
}

// call when data transformations are updated instead of generateTransformedOutput
func refilterTuiTable() {

	// record old selected entry
	// oldSelectedEntryIndex := -1
	// if len(outputEntryIndices) > 0 {
	// 	oldSelectedEntryIndex = outputEntryIndices[selR - data.numHeaderRows]
	// }

	// re-transform data
	transformDataToOutput()
	updateInfoText()

	// see if we can keep the same entry selected ( I disabled since it wasn't very useful and just annoying )
	// if newSelectedEntryIndexIndex := slices.Index(outputEntryIndices, oldSelectedEntryIndex); newSelectedEntryIndexIndex != -1 {
	// 	selR = newSelectedEntryIndexIndex + data.numHeaderRows
	// } else if selR > len(outputEntryIndices) + data.numHeaderRows {
	// 	selR = data.numHeaderRows
	// }
}

const lastPresetName = "last"
func exitTui() {

	// save "last" preset
	presetTransformations[lastPresetName] = deepCopyPreset(transformation)
	savePresetsToFile()

	app.Stop()
}

func deleteColumn(column int) {
	transformation.ColumnHeaders = slices.Delete(transformation.ColumnHeaders, column, column + 1)
	if (selC >= len(transformation.ColumnHeaders)) { selC = len(transformation.ColumnHeaders) - 1 }

	if len(transformation.ColumnHeaders) == 0 {
		writeToMessageBuffer("Only to find Gideon's bible...")
	}
}

func writeToMessageBuffer(message string) {
	messageBuffer.SetText(message)
}

func setInstructionsText(text string) {
	instructionsText.SetText(text)
}

// FLOATING WINDOWS =========================================================================

var menuNameToCancelFunc map[string]func() = make(map[string]func())
func createFloatingMenu(name string, root tview.Primitive, cancelFunc func()) {

	// if page exists, close it
	if pages.HasPage(name) {
		menuNameToCancelFunc[name]()
	}

	const width = 80
	const height = 30
	floating := tview.NewFlex().
	AddItem(nil, 0, 1, false).
	AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
	AddItem(nil, 0, 1, false).
	AddItem(root, height, 0, true).
	AddItem(nil, 0, 1, false), width, 0, true).
	AddItem(nil, 0, 1, false)

	floating.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyESC:
			cancelFunc()
			return nil
		}

		return event
	})

	menuNameToCancelFunc[name] = cancelFunc
	pages.AddPage(name, floating, true, true)
	app.SetFocus(root)
}

// ACTIONS ERRROR CHECKING ==================================================================
func confirmValidColumnSelection() bool {
	return len(transformation.ColumnHeaders) > 0
}

func confirmValidRowSelection() bool {
	return len(outputEntryIndices) > 0
}

// ACTIONS ==================================================================================

func moveColumn(delta int) {
	if !confirmValidColumnSelection() { return }

	// calculate new columns (go doesn't have clamp function?)
	newColumn := max(selC + delta, 0)
	if newColumn >= len(transformation.ColumnHeaders) { newColumn = len(transformation.ColumnHeaders) - 1 }

	if newColumn != selC {
		// swap columns
		transformation.ColumnHeaders[selC], transformation.ColumnHeaders[newColumn] = transformation.ColumnHeaders[newColumn], transformation.ColumnHeaders[selC] 
		selC = newColumn
	} 
}

const saveMenuPageName = "saveMenu"
func openSaveMenu() {
	form := tview.NewForm()
	form.SetTitle("Save Transformation").SetBorder(true)

	doneFunc := func()  {
		pages.RemovePage(saveMenuPageName)
	}

	// file name input
	var name string = activePresetName
	form.AddInputField("Name", name, 30, nil, func(text string) {
		name = text
	})

	// save as preset button
	form.AddButton("Save as preset", func() {

		if name == "" { return }

		// save to preset list
		presetTransformations[name] = deepCopyPreset(transformation)
		activePresetName = name

		// save preset list to file
		savePresetsToFile()

		doneFunc()
	})

	// save as file button
	form.AddButton("Save as file", func() {
		if name == "" { return }

		json, _ := serializeTransformation()
		err := os.WriteFile(name, json, 0666)
		if (err != nil) {
			log.Fatalf("Could not write saved transformation to file: %v", err)
		}

		writeToMessageBuffer(fmt.Sprintf("Saved preset to file %v", name))

		doneFunc()
	})

	// print button (and exit app)
	form.AddButton("Print to stdout", func() {
		exitTui()
		json, _ := serializeTransformation()
		fmt.Print(string(json))
	})

	// cancel button
	form.AddButton("Cancel", func() { 
		doneFunc()
	})

	createFloatingMenu(saveMenuPageName, form, doneFunc)
	setInstructionsText(getInstructionString([]string{"app", "floating"}))
}

const presetMenuPageName = "columnMenu"
func openPresetMenu() {
	list := tview.NewList()
	list.SetTitle("Preset Menu").SetBorder(true)

	doneFunc := func()  {
		pages.RemovePage(presetMenuPageName)
	}

	// quit button
	list.AddItem("quit", "", 'q', doneFunc)

	getPresetAltText := func(presetName string) string  {
		if presetName == lastPresetName {
			return "[orange::]Special Preset[w::]"
		}

		altText := "[red::]Press x to remove[w::]"
		if presetName == activePresetName { 
			altText = "[green::]Active. " + altText
		}
		return altText
	}

	// add each preset as list item
	activePresetIndex := -1
	index := 1
	for preset, _ := range presetTransformations {
		list.AddItem(preset, getPresetAltText(preset), 0, nil)
		if (preset == activePresetName) { activePresetIndex = index }

		index += 1
	}

	// use preset
	list.SetSelectedFunc(func(i int, presetName, alt string, r rune) {
		if i == 0 { return }

		// use preset
		oldPresetName := activePresetName
		usePreset(presetName)
		refilterTuiTable()

		// update list text
		list.SetItemText(i, presetName, getPresetAltText(presetName))
		if activePresetIndex > 0 { list.SetItemText(activePresetIndex, oldPresetName, getPresetAltText(oldPresetName)) }
		activePresetIndex = i

		writeToMessageBuffer(fmt.Sprintf("Using preset: %v", presetName))
	})

	// x to remove preset
	list.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyRune:
			if (event.Rune() == 'x') {
				index := list.GetCurrentItem()
				if index == 0 { doneFunc() }

				presetName, _ := list.GetItemText(index)
				if presetName == lastPresetName {
					return event
				}

				if index == activePresetIndex {
					activePresetIndex = -1
				}

				// update preset data
				delete(presetTransformations, presetName)
				savePresetsToFile()

				// update list
				list.RemoveItem(index)
			}
		}

		return event
	})

	createFloatingMenu(presetMenuPageName, list, doneFunc)
	setInstructionsText(getInstructionString([]string{"app", "floating"}))
}

const columnMenuPageName = "columnMenu"
func openColumnMenu() {
	list := tview.NewList()
	list.SetTitle("Column Menu").SetBorder(true)

	doneFunc := func()  {
		pages.RemovePage(columnMenuPageName)
	}

	// quit button
	list.AddItem("quit", "", 'q', doneFunc)

	// gets secondary text for each object
	isHeaderActive := func(header string) bool {
		return slices.Contains(transformation.ColumnHeaders, header)
	}
	getHeaderAltText := func(header string) string {
		if isHeaderActive(header) {
			return "active"
		} else {
			return "[red::]inactive[w::]"
		}
	}

	// add each header as list item
	for _, header := range data.columnHeaders {
		list.AddItem(header, getHeaderAltText(header), 0, nil)
	}
	for _, header := range transformation.ColumnHeaders {
		if !isColumnFake(header) { continue }
		list.AddItem(header, "[orange::]Fake column[w::]", 0, nil) 
	}
	// remove/add header
	list.SetSelectedFunc(func(i int, header, alt string, r rune) {
		if i == 0 { return }

		// update transformation
		if isHeaderActive(header) {
			deleteColumn(slices.Index(transformation.ColumnHeaders, header))
		} else {
			transformation.ColumnHeaders = append(transformation.ColumnHeaders, header)
		}

		// update list
		list.RemoveItem(i)
		if (!isColumnFake(header)) {
			list.InsertItem(i, header, getHeaderAltText(header), r, nil)
			list.SetCurrentItem(i)
		}

		// re-render
		refilterTuiTable()
	})

	// TODO - displays headers not in data that exist in transformation

	createFloatingMenu(columnMenuPageName, list, doneFunc)
	setInstructionsText(getInstructionString([]string{"app", "floating"}))
}

const filterMenuPageName = "filterMenu"
func openColumnFilterMenu() {
	if !confirmValidColumnSelection() { return }

	columnHeader := transformation.ColumnHeaders[selC]

	var includeRegex string
	if includeVal, includeFound := transformation.IncludeRegexByColumn[columnHeader]; includeFound {
		includeRegex = includeVal
	}

	var excludeRegex string
	if excludeVal, excludeFound := transformation.ExcludeRegexByColumn[columnHeader]; excludeFound {
		excludeRegex = excludeVal
	}

	// regex inputs
	filterMenu := tview.NewForm() 
	filterMenu.SetBorder(true).SetTitle("Filter Menu")
	// include
	filterMenu.AddInputField("Include Regex", includeRegex, 50, nil, func(text string) {
		includeRegex = text
	})
	// exclude
	filterMenu.AddInputField("Exclude Regex", excludeRegex, 50, nil, func(text string) {
		excludeRegex = text
	})

	// finish function
	finishFunc := func() {

		if includeRegex != "" {
			transformation.IncludeRegexByColumn[columnHeader] = includeRegex
		} else {
			delete(transformation.IncludeRegexByColumn, columnHeader)
		}

		if excludeRegex != "" {
			transformation.ExcludeRegexByColumn[columnHeader] = excludeRegex
		} else {
			delete(transformation.ExcludeRegexByColumn, columnHeader)
		}

		pages.RemovePage(filterMenuPageName)
		refilterTuiTable()
	}

	cancelFunc := func() {
		pages.RemovePage(filterMenuPageName)
	}

	// exit methods
	filterMenu.AddButton("Done", finishFunc)
	filterMenu.AddButton("Cancel", cancelFunc)

	createFloatingMenu(filterMenuPageName, filterMenu, cancelFunc)
	setInstructionsText(getInstructionString([]string{"app", "floating"}))
}

func deleteSelectedColumn() {
	if !confirmValidColumnSelection() { return }

	deleteColumn(selC)

	refilterTuiTable()
}

func sortColumn() {
	if !confirmValidColumnSelection() { return }

	// update sort config
	newColumn := transformation.ColumnHeaders[selC]
	if newColumn != transformation.SortByColumn {
		transformation.SortByColumn = newColumn
	} else {
		transformation.SortAscending = !transformation.SortAscending
	}

	refilterTuiTable()
}

