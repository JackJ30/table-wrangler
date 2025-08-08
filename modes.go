package main

import (
	"fmt"
	"slices"
	"strings"

	"golang.design/x/clipboard"

	// external
	"github.com/gdamore/tcell/v2"
)

type TableMode interface {
	onEnter()
	onExit()
	getCellSelectionStatus(row int, column int) CellSelectStatus
	onSelected(row int, column int)
	onClicked(row int, column int) bool
	inputCapture(event *tcell.EventKey) *tcell.EventKey
	getInstructionCategories() []string
}

var tableModeStack []TableMode

func appendMode(mode TableMode) {

	if (len(tableModeStack) > 0) {
		tableModeStack[len(tableModeStack) - 1].onExit()
	}

	tableModeStack = append(tableModeStack, mode)
	tableModeStack[len(tableModeStack) - 1].onEnter()
	setCurrentModeInstructionText()
}

func dropMode() {
	tableModeStack[len(tableModeStack) - 1].onExit()
	tableModeStack = slices.Delete(tableModeStack, len(tableModeStack) - 1, len(tableModeStack))

	if len(tableModeStack) > 0 {
		tableModeStack[len(tableModeStack) - 1].onEnter()
		setCurrentModeInstructionText()
	}
}

func setMode(mode TableMode) {
	if (len(tableModeStack) > 0) {
		tableModeStack[len(tableModeStack) - 1].onExit()
	} else {
		tableModeStack = make([]TableMode, 1)
	}

	tableModeStack[len(tableModeStack) - 1] = mode
	tableModeStack[len(tableModeStack) - 1].onEnter()
	setCurrentModeInstructionText()
}

func setCurrentModeInstructionText() {
	setInstructionsText(getInstructionString(tableModeStack[len(tableModeStack) - 1].getInstructionCategories()))
}

// SELECTION MODES ====================================================================

func selectionModeInputCapture(event *tcell.EventKey) *tcell.EventKey {

	switch event.Key() {
	case tcell.KeyRune:
		if event.Rune() == 'q' {
			exitTui()
			return nil
		}
		if event.Rune() == 'c' {
			appendMode(CopyMode{})
			return nil
		}
		if event.Rune() == 'b' {
			appendMode(&BoxMode{})
			return nil
		}
		if event.Rune() == 'p' {
			exitTui()
			printTable()
			return nil
		}
	}

	if event.Modifiers()&tcell.ModCtrl != 0 {
		switch event.Key() {
		case tcell.KeyCtrlS:
			openSaveMenu()
		case tcell.KeyCtrlY:
			openColumnMenu()
		case tcell.KeyCtrlP:
			openPresetMenu()
		}
	}

	return event
}

// COLUMN MODE ========================================================================

type ColumnMode struct {

}

func (m ColumnMode) onEnter() {
}

func (m ColumnMode) onExit() {
}

func (m ColumnMode) getCellSelectionStatus(row int, column int) CellSelectStatus {
	if column == selC {
		return SelectionPrimary
	} else if row == selR {
		return SelectionSecondary
	} else {
		return SelectionNone
	}
}

func (m ColumnMode) onSelected(row int, column int) {
	selC = column
}

func (m ColumnMode) onClicked(row int, column int) bool {
	return false
}

func (m ColumnMode) inputCapture(event *tcell.EventKey) *tcell.EventKey {
	if selectionModeInputCapture(event) == nil {
		return nil
	}

	switch event.Key() {
	case tcell.KeyRune:
		if event.Rune() == 'v' {
			setMode(RowMode{})
			return nil
		}
		if event.Rune() == 'x' {
			deleteSelectedColumn()
			return nil
		}
		if event.Rune() == 's' {
			sortColumn()
			return nil
		}
		if event.Rune() == 'f' {
			openColumnFilterMenu()
			return nil
		}
	}
	if event.Modifiers()&tcell.ModCtrl != 0 {
		switch event.Key() {
		case tcell.KeyCtrlQ:
			moveColumn(-1)
		case tcell.KeyCtrlE:
			moveColumn(1)
		}
	}

	return event
}

func (m ColumnMode) getInstructionCategories() []string {
	return []string{ "app", "selection", "column" }
}

// ROW MODE ===========================================================================

type RowMode struct {

}

func (m RowMode) onEnter() {
}

func (m RowMode) onExit() {
}

func (m RowMode) getCellSelectionStatus(row int, column int) CellSelectStatus {
	if row == selR {
		return SelectionPrimary
	} else if column == selC {
		return SelectionSecondary
	} else {
		return SelectionNone
	}
}

func (m RowMode) onSelected(row int, column int) {
	selR = row
}

func (m RowMode) onClicked(row int, column int) bool {
	return false
}

func (m RowMode) inputCapture(event *tcell.EventKey) *tcell.EventKey {
	if selectionModeInputCapture(event) == nil {
		return nil
	}

	switch event.Key() {
	case tcell.KeyRune:
		if event.Rune() == 'v' {
			setMode(ColumnMode{})
			return nil
		}
	}

	return event
}

func (m RowMode) getInstructionCategories() []string {
	return []string{ "app", "selection" }
}

// Copy Mode ====================================================================

type CopyMode struct {

}

func (m CopyMode) onEnter() {
}

func (m CopyMode) onExit() {
}

func getDelimitedString(slice []string) string {
	return strings.Join(slice, " ")
}

func (m CopyMode) copy(row int, column int) {
	clipboard.Write(clipboard.FmtText, []byte(table.GetCell(row, column).Text))
	writeToMessageBuffer(fmt.Sprintf("Copied text at (%v, %v)", row, column))
	dropMode()
}

func (m CopyMode) copyColumn(column int) {
	var output []string
	for r := data.numHeaderRows; r < table.GetRowCount(); r += 1 {
		output = append(output, table.GetCell(r, column).Text)
	}
	clipboard.Write(clipboard.FmtText, []byte(getDelimitedString(output)))
	writeToMessageBuffer(fmt.Sprintf("Copied column %v", column))
	dropMode()
}

func (m CopyMode) getCellSelectionStatus(row int, column int) CellSelectStatus {
	if row == selR && column == selC {
		return SelectionPrimary
	} else if row == selR {
		return SelectionSecondary
	} else if column == selC {
		return SelectionSecondary
	} else {
		return SelectionNone
	}
}

func (m CopyMode) onSelected(row int, column int) {
	selR = row
	selC = column
}

func (m CopyMode) onClicked(row int, column int) bool {

	m.copy(row, column)

	return true
}

func (m CopyMode) inputCapture(event *tcell.EventKey) *tcell.EventKey {

	switch event.Key() {
	case tcell.KeyRune:
		if event.Rune() == 'q' {
			dropMode()
			return nil
		}
		if event.Rune() == 'c' {
			m.copy(selR, selC)
			return nil
		}
		if event.Rune() == 'C' {
			m.copyColumn(selC)
			return nil
		}
		if event.Rune() == 'v' {
			dropMode()
			return nil
		}
	case tcell.KeyEsc:
		dropMode()
		return nil
	}

	return event
}

func (m CopyMode) getInstructionCategories() []string {
	return []string{ "app", "copy" }
}

// Box Mode ====================================================================

type BoxMode struct {
	markR, markC int
}

func (m *BoxMode) getBounds() (minR, maxR, minC, maxC int) {
	return min(m.markR, selR), max(m.markR, selR), min(m.markC, selC), max(m.markC, selC)
}

func (m *BoxMode) onEnter() {
	m.markR, m.markC = selR, selC
}

func (m *BoxMode) onExit() {
}

func (m *BoxMode) getCellSelectionStatus(row int, column int) CellSelectStatus {
	if row == selR && column == selC {
		return SelectionPoint
	}

	minR, maxR, minC, maxC := m.getBounds()
	if (row >= minR && row <= maxR) && (column >= minC && column <= maxC) {
		return SelectionPrimary
	}

	return SelectionNone
}

func (m *BoxMode) onSelected(row int, column int) {
	selR = row
	selC = column
}

func (m *BoxMode) onClicked(row int, column int) bool {
	return true
}

func (m *BoxMode) inputCapture(event *tcell.EventKey) *tcell.EventKey {

	switch event.Key() {
	case tcell.KeyRune:
		if event.Rune() == 'q' {
			dropMode()
			return nil
		}
		if event.Rune() == 'c' {
			minR, maxR, minC, maxC := m.getBounds()

			// copy in selection
			var output []string
			for c := minC; c <= maxC; c += 1 {
				for r := minR; r <= maxR; r += 1 {
					output = append(output, table.GetCell(r, c).Text)
				}
			}

			clipboard.Write(clipboard.FmtText, []byte(getDelimitedString(output)))
			writeToMessageBuffer("Copied text")
			dropMode()

			return nil
		}
	case tcell.KeyEsc:
		dropMode()
		return nil
	}

	return event
}

func (m *BoxMode) getInstructionCategories() []string {
	return []string{ "app", "box" }
}
