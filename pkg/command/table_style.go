package command

import "github.com/jedib0t/go-pretty/v6/table"

var DefaultTableStyle = table.Style{
	Name:    "Default",
	Box:     table.StyleBoxDefault,
	Color:   table.ColorOptionsDefault,
	Format:  table.FormatOptionsDefault,
	HTML:    table.DefaultHTMLOptions,
	Options: table.OptionsNoBordersAndSeparators,
	Title:   table.TitleOptionsDefault,
}
