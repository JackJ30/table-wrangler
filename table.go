package main

import (
	"github.com/gdamore/tcell/v2"
	"github.com/JackJ30/jack-tview"
)

// live data
var selC, selR int = 0, 0
var offsetC, offsetR = 0, 0

var tableData *TableData

func createTable() {

	// init live data
	selC = 0
	selR = data.numHeaderRows

	// create table object
	table = tview.NewTable().SetBorders(*flags.fatTable).SetEvaluateAllRows(true).SetFixed(data.numHeaderRows, 0).SetSelectable(false, false)
	table.SetBorder(false)

	//set data
	tableData = &TableData{}
	table.SetContent(tableData)

	// input handling
	table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		// process table controls
		response := func(event *tcell.EventKey) *tcell.EventKey  {
			switch event.Key() {
			case tcell.KeyRune:
				switch event.Rune() {
				case 'g':
					selR = data.numHeaderRows
					return nil
				case 'G':
					selR = tableData.GetRowCount() - 1
					return nil
				case 'j':
					selR += 1
					return nil
				case 'k':
					selR -= 1
					return nil
				case 'h':
					selC -= 1
					return nil
				case 'l':
					selC += 1
					return nil
				}
			case tcell.KeyHome:
				selR = data.numHeaderRows
				return nil
			case tcell.KeyEnd:
				selR = tableData.GetRowCount() - 1
				return nil
			case tcell.KeyUp:
				selR -= 1
				return nil
			case tcell.KeyDown:
				selR += 1
				return nil
			case tcell.KeyLeft:
				selC -= 1
				return nil
			case tcell.KeyRight:
				selC += 1
				return nil
			}
			return event
		}(event)
		
		// clamp selection
		clampRowToTable := func(row int, includeHeader bool) int {
			if row < data.numHeaderRows && includeHeader {
				return data.numHeaderRows
			} else if row < 0 && !includeHeader {
				return 0
			} else if row >= tableData.GetRowCount() {
				return tableData.GetRowCount() - 1
			} else {
				return row
			}
		}
		clampColumnToTable := func(column int) int {
			if column < 0 {
				return 0
			} else if column >= tableData.GetColumnCount() {
				return tableData.GetColumnCount() - 1
			} else {
				return column
			}
		}
		selR = clampRowToTable(selR, true)
		selC = clampColumnToTable(selC)


		const scrollPaddingC = 2
		// move columns offset left
		if selC < offsetC + scrollPaddingC { offsetC = selC - scrollPaddingC }
		// move columns offset top
		dif := selC - (offsetC + len(table.GetVisibleColumnIndices()) - scrollPaddingC)
		if dif >= 0 { offsetC += dif }

		const scrollPaddingR = 4
		// move rows offset bottom
		if selR < offsetR + scrollPaddingR { offsetR = selR - scrollPaddingR }
		// move rows offset top
		dif = selR - (offsetR + table.GetVisibleRowCount() - scrollPaddingR)
		if dif >= 0 { offsetR += dif }

		// clamp offsets and set
		offsetR = clampRowToTable(offsetR, false)
		offsetC = clampColumnToTable(offsetC)
		table.SetOffset(offsetR, offsetC)

		// otherwise send input to mode
		if (response == nil) { return response }
		return tableModeStack[len(tableModeStack) - 1].inputCapture(event)
	})

	// automatically close floating windows and set mode text on focus
	table.SetFocusFunc(func() {
		for _, pageName := range pages.GetPageNames(false) {
			if pageName != "main" {
				// run cancel func
				menuNameToCancelFunc[pageName]()
			}
		}
		setCurrentModeInstructionText()
	})
}

type TableData struct {
	tview.TableContentReadOnly
}

func (d *TableData) GetCell(row, column int) *tview.TableCell {
	var cell *tview.TableCell = nil

	if column == -1 || row == -1 {
		return nil
	}

	header := transformation.ColumnHeaders[column]
	fake := isColumnFake(header)
	selectionMode := tableModeStack[len(tableModeStack) - 1].getCellSelectionStatus(row, column)

	if row < data.numHeaderRows {
		// if header
		content := decorateHeader(header)
		cell = colorizeTCell(tview.NewTableCell(content).SetAlign(tview.AlignCenter), true, column, fake, false, selectionMode)
	} else if len(outputEntryIndices) == 0 {
		// if "empty" entry
		cell = colorizeTCell(tview.NewTableCell("EMPTY").SetAlign(tview.AlignCenter).SetClickedFunc(func() bool {
			return tableModeStack[len(tableModeStack) - 1].onClicked(row, column)
		}), false, column, fake, true, selectionMode)
	} else {
		// if entry
		content := getDataInColumn(header, outputEntryIndices[row - data.numHeaderRows])
		cell = colorizeTCell(tview.NewTableCell(content).SetAlign(tview.AlignCenter).SetClickedFunc(func() bool {
			return tableModeStack[len(tableModeStack) - 1].onClicked(row, column)
		}), false, column, fake, false, selectionMode)
	}

	return cell
}

func (d *TableData) GetRowCount() int {
	// if empty, 1 column for "EMPTY"
	if len(outputEntryIndices) == 0 {
		return 1 + data.numHeaderRows
	}

	return len(outputEntryIndices) + data.numHeaderRows
}

func (d *TableData) GetColumnCount() int {
	return len(transformation.ColumnHeaders)
}

