package tableview

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

func tviewFillTable(table *tview.Table, columns []string, data [][]string) {
	for i := 0; i < len(columns); i++ {
		cell := tview.NewTableCell("[yellow]" + columns[i]).SetBackgroundColor(tcell.ColorBlue)
		cell.SetSelectable(false)
		table.SetCell(0, i, cell)
		for j := 0; j < len(data); j++ {
			content := data[j][i]
			cell := tview.NewTableCell(content)
			cell.SetMaxWidth(32)
			table.SetCell(j+1, i, cell)
		}
	}
}

type tableViewCommand struct {
	ch     rune
	text   string
	action func(row int)
}

type TableView struct {
	columns  []string
	data     [][]string
	commands []tableViewCommand
	table    *tview.Table
}

func NewTableView() *TableView {
	t := new(TableView)
	t.table = tview.NewTable()
	return t
}

func (t *TableView) FillTable(columns []string, data [][]string) {
	tviewFillTable(t.table, columns, data)
}

func (t *TableView) NewRow() {
}

func (t *TableView) NewColumn() {
}

func (t *TableView) NewCommand(ch rune, text string, action func(row int)) {
	t.commands = append(t.commands, tableViewCommand{ch, text, action})
}

func (t *TableView) DelRow() {
}

func (t *TableView) DelColumn() {
}

func (t *TableView) Run() {
	app := tview.NewApplication()
	text := tview.NewTextView()
	flex := tview.NewFlex()
	var lastLine tview.Primitive
	var lastSearch string

	tviewSearch := func(row int, text string) bool {
		text = strings.ToLower(text)
		for i := 0; i < len(t.data); i++ {
			for j := 0; j < len(t.columns); j++ {
				cellContent := strings.ToLower(t.data[(row+i)%len(t.data)][j])
				if strings.Contains(cellContent, text) {
					t.table.Select(((row+i)%len(t.data))+1, 0)
					return true
				}
			}
		}
		return false
	}

	// t.table.SetBorder(true)
	t.table.SetTitle(" LDAP ")
	t.table.SetFixedColumnsWidth(true)
	// t.table.SetBorders(true)
	t.table.SetSeparator(tview.Borders.Vertical)
	t.table.SetFixed(1, 0)
	t.table.SetSelectable(true, false)
	tviewFillTable(t.table, t.columns, t.data)
	t.table.SetDoneFunc(func(key tcell.Key) {
		app.Stop()
	})
	t.table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyRune:
			switch event.Rune() {
			case 'q':
				app.Stop()
				return nil
			case '/':
				row, _ := t.table.GetSelection()
				row--
				search := tview.NewInputField()
				search.SetLabel("Search: ")
				search.SetFieldBackgroundColor((tcell.ColorBlack))
				search.SetChangedFunc(func(text string) {
					if tviewSearch(row, text) {
						search.SetFieldTextColor(tcell.ColorWhite)
					} else {
						search.SetFieldTextColor((tcell.ColorRed))
					}
				})
				search.SetDoneFunc(func(key tcell.Key) {
					lastSearch = search.GetText()
					flex.RemoveItem(lastLine)
					lastLine = tview.NewTextView().SetText(fmt.Sprintf("Last search: %q from line %d", lastSearch, row))
					flex.AddItem(lastLine, 1, 0, false)
					app.SetFocus(t.table)
				})
				flex.RemoveItem(lastLine)
				lastLine = search
				flex.AddItem(lastLine, 1, 0, false)
				app.SetFocus(search)

			case 'n':
				row, _ := t.table.GetSelection()
				tviewSearch(row, lastSearch)
				flex.RemoveItem(lastLine)
				lastLine = tview.NewTextView().SetText(fmt.Sprintf("Searching again: %q from line %d", lastSearch, row))
				flex.AddItem(lastLine, 1, 0, false)
			}
			for _, c := range t.commands {
				if event.Rune() == c.ch {
					row, _ := t.table.GetSelection()
					app.Suspend(func() {
						c.action(row)
					})
				}
			}
		}
		return event
	})
	text.SetBackgroundColor(tcell.ColorBlue)
	text.SetDynamicColors(true)
	innerText := " [yellow]q:quit   /:search   n:next"
	for _, c := range t.commands {
		innerText = fmt.Sprintf("%s   %c:%s", innerText, c.ch, c.text)
	}
	text.SetText(innerText)
	flex.SetBackgroundColor(tcell.ColorRed)
	flex.SetDirection(tview.FlexRow)
	flex.AddItem(t.table, 0, 1, true)
	flex.AddItem(text, 1, 0, false)
	lastLine = tview.NewBox()
	flex.AddItem(lastLine, 1, 0, false)
	app.SetRoot(flex, true)
	if err := app.Run(); err != nil {
		panic(err)
	}
}
