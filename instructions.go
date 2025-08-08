package main

import "fmt"

type InstructionCategory struct {
	header string
	description string
	instructions []string
}

func getInstructionString(categories []string) string {
	instructions := ""

	// Add each category
	for _, categoryName := range categories {
		category := instructionCategories[categoryName]
		instructions += fmt.Sprintf("[purple::b]%v[w::-]\n", category.header)
		if category.description != "" { instructions += category.description + "\n" }
		for _, instruction := range category.instructions {
			instructions += instruction + "\n"
		}
		instructions += "\n"
	}

	return instructions
}

var instructionCategories map[string]InstructionCategory = map[string]InstructionCategory{
	"app" : {
		header: "App Instructions",
		description: "",
		instructions: []string{
			"[::b]C-c[::-] - quit.",
			"[::b]C-h[::-] - toggle this panel.",
		},
	},
	"selection" : {
		header: "General Instructions",
		description: "",
		instructions: []string{
			"[::b]v[::-] - switch selection mode",
			"[::b]c[::-] - copy mode.",
			"[::b]b[::-] - box mode.",
			"[::b]C-s[::-] - open save menu.",
			"[::b]C-p[::-] - open preset menu.",
			"[::b]C-y[::-] - open column menu.",
			"[::b]q[::-] - quit.",
			"[::b]p[::-] - quit and print.",
		},
	},
	"column" : {
		header: "Column Mode Instructions",
		description: "",
		instructions: []string{
			"[::b]x[::-] - delete column.",
			"[::b]s[::-] - sort by column.",
			"[::b]f[::-] - filter column.",
			"[::b]C-q[::-] - move column left.",
			"[::b]C-e[::-] - move column right.",
		},
	},
	"copy" : {
		header: "Copy Mode Instructions",
		description: "",
		instructions: []string{
			"[::b]click[::-] - copy cell at mouse.",
			"[::b]c[::-] - copy selected cell.",
			"[::b]C[::-] - copy selected column.",
			"[::b]esc, q[::-] - exit.",
		},
	},
	"box" : {
		header: "Box Select Mode Instructions",
		description: "",
		instructions: []string{
			"[::b]esc, q[::-] - exit.",
			"[::b]c[::-] - copy.",
		},
	},
	"floating" : {
		header: "Floating Window Instructions",
		description: "",
		instructions: []string{
			"[::b]tab[::-] - next.",
			"[::b]shift-tab[::-] - prev.",
			"[::b]return[::-] - select.",
			"[::b]esc[::-] - exit.",
		},
	},
}
