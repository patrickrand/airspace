package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/concourse/fly/ui"
	"github.com/fatih/color"
)

func main() {
	for range time.Tick(2 * time.Second) {
		if err := printBuilds("mia", "", 20); err != nil {
			panic(err)
		}
	}
}

func getBuilds(target, pattern string, count int) error {
	_, err := exec.Command("fly", "-t", target, "builds").CombinedOutput()
	if err != nil {
		return err
	}

	return nil
}

func printBuilds(target, pattern string, count int) error {
	output, err := exec.Command("fly", "-t", target, "builds", "-c", strconv.Itoa(count)).Output()
	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(bytes.NewReader(output))

	scanner.Scan()
	//header := scanner.Text()

	table := NewTable()
	for scanner.Scan() {
		build := scanner.Text()
		cols := strings.Fields(build)
		if len(cols) != 7 {
			return errors.New("unable to parse line: " + build)
		}
		table.Append(
			NewCell(cols[0]),
			NewCell(cols[1]),
			NewCell(cols[2]),
			NewCell(colorize(cols[3])),
			NewCell(cols[4]),
			NewCell(cols[5]),
			NewCell(cols[6]),
		)
	}

	buf := new(bytes.Buffer)
	table.Render(buf)
	fmt.Printf("\r%s", buf.String())
	fmt.Printf("\033[s\033[" + strconv.Itoa(count) + "A")
	return nil
}

func colorize(status string) string {
	switch status {
	case "succeeded":
		return color.GreenString(status)
	case "failed":
		return color.RedString(status)
	case "aborted":
		return color.MagentaString(status)
	case "errored":
		return color.New(color.FgWhite, color.BgRed).SprintFunc()(status)
	default:
		return status
	}
}

type Table struct {
	ui.Table
}

type Cell struct {
	ui.TableCell
}

func NewCell(contents string) *Cell {
	return &Cell{ui.TableCell{Contents: contents}}
}

func NewTable() *Table {
	return &Table{ui.Table{
		Headers: ui.TableRow{
			{Contents: "id", Color: color.New(color.Bold)},
			{Contents: "pipeline/job", Color: color.New(color.Bold)},
			{Contents: "build", Color: color.New(color.Bold)},
			{Contents: "status", Color: color.New(color.Bold)},
			{Contents: "start", Color: color.New(color.Bold)},
			{Contents: "end", Color: color.New(color.Bold)},
			{Contents: "duration", Color: color.New(color.Bold)},
		},
	}}
}

func (t *Table) Append(cells ...*Cell) {
	var tableCells []ui.TableCell
	for _, c := range cells {
		tableCells = append(tableCells, c.TableCell)
	}
	t.Data = append(t.Data, tableCells)
}

func (t *Table) Render(w io.Writer) error {
	return t.Table.Render(os.Stdout)
}
