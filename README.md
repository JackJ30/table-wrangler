# Jack's Table Wrangler
This tool displays and transforms tables

There are two use cases:
- Use the TUI as a table viewer and transformer
- Create transformation presets, apply them on the command line without leaving your shell

Notes:
- Read the [tutorial](#Tutorial)
- I tried to keep the code fairly clean, and there is still some jank, but I think it is surprisingly robust.
- Check out all of the options with `-help`, and try the keybinds on the instruction panel.
- I tried to design the program to integrate well into existing terminal workflows. You can use the tui once, save a preset and then never enter it again (use `stdout` flag). It will automatically detect if it is being piped into other programs.

## Building and Installing
1. Clone the repo and change directory to it
2. Get the dependencies with `go get .`
3. Build with `go build`

## Usage Examples
- `cmd | table-wrangler`
- `table-wrangler -command="cmd"`
- `table-wrangler -p="presetName"`

## Tutorial
Start with a commmand that outputs a table. We're going to filter some entries and make the table smaller.

### Basics
Pipe the command into `table-wrangler`, this will open the tui as soon as the command finishes. Here are the basics:
- All keybinds are displayed on the control panel on the right.
- You can navigate the table view with arrow keys or vim keys.
- There are two main "selection modes": row and column. Press **v** to switch between them. Column mode has additional controls.
- You can exit the program with **q** or **C-c**.

### Deleting Columns
Let's make the table smaller (horizontally). Go into column mode and press **x** to remove the column. If you accidentally delete the column, you can revive it in the column menu opened with **C-y**. Each menu and mode have their own keybinds which you can read from the control panel.

### Filtering Columns
Let's add some filters. Select a column in column mode, and press **f** to open the filter menu. You can use basic regex like `value|other-value`.

### Sorting Columns
If you want to sort a column (so that the same values appear next to each other), press **s** in column mode.

### Reordering Columns
If you want to change the order of columns, you can use **C-q** and **C-e** to move columns left and right.

### Copying Data
- To get data out of the TUI, press **c** to enter copy mode. Click on a cell to copy its contents.
- You can box select by pressing **b**, and then copy the box selected contents by pressing **c**.

### Using without the TUI
Once you have transformed your data, you might find it annoying that the TUI locks down your terminal until you exit. Here are a few other ways you can use `table-wrangler'.
- Press **p** in the TUI to exit and print the table to stdout.
- Save a preset by opening the save menu with **C-s**, naming it, and selecting "Save as preset". You can specify a preset on the command line with 'table-wrangler -p=presetName'. You can also load presets from the TUI in the preset menu (opened with **C-p**).
- Use the `-stdout` flag to skip the tui and immediately print to stdout (useful when paired with a preset).
- Use the special "last" preset which is automatically saved whenever you exit the TUI.

### Tip
Make sure to read the keybinds in the control panel for each mode, and check the command line arguments with `table-wrangler -h`.

## TODO
Small Improvements
- [ ] Input validation and error checking
- [ ] Colorize based on unique values in column
- [ ] Filtering completion based on possible values
- [ ] Custom header alias
- [ ] Show unfiltered data toggle
- [ ] Refresh command

### OLD

UI
- [x] Column ordering
- [x] Concept of a selected row and column
- [x] Click to copyable text
- [x] Selections stay trough re-filters
- [x] Mode system
- [x] Better copy mode
- [x] Better floating window system
- [x] Toggle instructions panel
- [x] Column editor
- [x] Preset editor
- [x] Sectioned instructions and floating window instructions
- [x] Handle "incompatible" presets
- [x] Remake selection system with better graphics
- [x] Box selection (mass copy)
- [x] Style (better colors)

TRANSFORMATIONS
- [x] Column hiding
- [x] Row sorting
- [x] Column re-ordering
- [x] filter by row value

UTIL
- [x] Intelligent position parse
- [x] Export to stdout instead of tui
- [x] Read from stdin instead of command
- [x] Saved transformations
- [x] Config presets

ARCHITECTURE
- [x] Make it work with no columns/rows + unify row/columns ops to have more consistent error checking

FIXES
- [x] Fix selection resetting
