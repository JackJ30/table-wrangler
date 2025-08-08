package main

import (
	"github.com/gdamore/tcell/v2"
	"github.com/JackJ30/jack-tview"
)

type CellSelectStatus int
const (
	SelectionPoint CellSelectStatus = iota
	SelectionPrimary
	SelectionSecondary
	SelectionNone
)

func colorizeTCell(cell *tview.TableCell, isHeader bool, column int, fake bool, empty bool, selection CellSelectStatus) *tview.TableCell {
	var backgroundColor tcell.Color = tcell.ColorDefault
	var textColor = tcell.ColorDefault

	// base colors
	if isHeader {
		// header colors
		if fake {
			backgroundColor = tcell.ColorDarkRed
		} else if column % 2 == 0 {
			backgroundColor = tcell.ColorDarkGreen
		} else {
			backgroundColor = tcell.ColorDarkBlue
		}
	} else {
		// text color
		if empty {
			backgroundColor = tcell.ColorGrey
		} else if fake {
			backgroundColor = tcell.ColorRed
		} else if column % 2 == 0 {
			textColor = tcell.ColorGreen
		} else {
			textColor = tcell.ColorLightBlue
		}
	}

	// insane selection logic
	if !isHeader {
		switch selection {
		case SelectionPrimary:
			if fake || empty {
				textColor, backgroundColor = backgroundColor, tcell.ColorWhite
			} else {
				textColor, backgroundColor = tcell.ColorBlack, textColor
			}
		case SelectionSecondary:
			if !fake {
				backgroundColor = tcell.Color239
			}
		}
	}
	if selection == SelectionPoint {
		backgroundColor, textColor = tcell.ColorWhite, tcell.ColorBlack
	}

	cell.SetBackgroundColor(backgroundColor)
	cell.SetTextColor(textColor)

	return cell
}

func colorizeAnsiCell(value string, isHeader bool, column int, fake bool) string {
	if isHeader {
		if fake {
			value = "\033[48;5;1m" + value
		} else if column % 2 == 0 {
			value = "\033[48;5;22m" + value
		} else {
			value = "\033[48;5;18m" + value
		}
	} else {
		if fake {
			value = "\033[48;5;9m" + value
		} else if column % 2 == 0 {
			value = "\033[38;5;46m" + value
		} else {
			value = "\033[38;5;117m" + value
		}
	}

	return value + "\033[0m"
}

func decorateHeader(header string) string {
	decoratedHeader := header

	if transformation.SortByColumn == header {
		if transformation.SortAscending {
			decoratedHeader += "(↑)"
		} else {
			decoratedHeader += "(↓)"
		}
	}

	if _, includeFound := transformation.IncludeRegexByColumn[header]; includeFound {
		decoratedHeader += "(F)"
	} else if _, excludeFound := transformation.ExcludeRegexByColumn[header]; excludeFound {
		decoratedHeader += "(F)"
	}

	return decoratedHeader
}
