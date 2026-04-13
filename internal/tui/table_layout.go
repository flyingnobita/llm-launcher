package tui

import (
	"github.com/mattn/go-runewidth"

	"github.com/flyingnobita/llm-launch/internal/llamacpp"
	btable "github.com/flyingnobita/llm-launch/internal/tui/btable"
)

func tableColumns(totalWidth int, files []llamacpp.ModelFile) []btable.Column {
	if totalWidth < 56 {
		totalWidth = 56
	}
	nameW, sizeW, modW, paramW := 36, 9, 17, 18
	longestName := 0
	longestPath := 0
	for _, f := range files {
		if w := runewidth.StringWidth(f.Name); w > longestName {
			longestName = w
		}
		if w := runewidth.StringWidth(llamacpp.FormatModelFolderDisplay(f.Path)); w > longestPath {
			longestPath = w
		}
	}
	if longestName > nameW {
		nameW = longestName
		if nameW > 72 {
			nameW = 72
		}
	}
	fixed := nameW + sizeW + modW + paramW + 8
	pathW := totalWidth - fixed
	if pathW < 14 {
		pathW = 14
	}
	if longestPath+2 > pathW {
		pathW = longestPath + 2
	}
	if pathW > 400 {
		pathW = 400
	}

	return []btable.Column{
		{Title: "Name", Width: nameW},
		{Title: "Path", Width: pathW},
		{Title: "Size", Width: sizeW},
		{Title: "Last modified", Width: modW},
		{Title: "Parameters", Width: paramW},
	}
}

// tableContentMinWidth approximates one row width (bubbles table pads cells) for setting table viewport width.
func tableContentMinWidth(cols []btable.Column) int {
	sum := 0
	for _, c := range cols {
		sum += c.Width
	}
	return sum + 4*len(cols)
}

func buildTableRows(files []llamacpp.ModelFile, cols []btable.Column) []btable.Row {
	if len(cols) < 5 {
		return nil
	}
	rows := make([]btable.Row, len(files))
	for i, f := range files {
		rows[i] = btable.Row{
			llamacpp.TruncateRunes(f.Name, cols[0].Width-1),
			llamacpp.TruncateRunes(llamacpp.FormatModelFolderDisplay(f.Path), cols[1].Width-1),
			llamacpp.FormatSize(f.Size),
			llamacpp.FormatModTime(f.ModTime),
			llamacpp.TruncateRunes(f.Parameters, cols[4].Width-1),
		}
	}
	return rows
}
