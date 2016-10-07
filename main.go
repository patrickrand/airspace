package main

import (
	"bufio"
	"bytes"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/concourse/fly/ui"
	"github.com/fatih/color"
)

var (
	countFlag       = flag.Int("c", 10, "count")
	targetFlag      = flag.String("t", "local", "target")
	pipelineJobFlag = flag.String("pj", "", "pipeline/job regex")
)

func main() {
	flag.Parse()
	for range time.Tick(2 * time.Second) {
		if err := run(*targetFlag, *pipelineJobFlag, *countFlag); err != nil {
			panic(err)
		}
	}
}

func run(target, pattern string, count int) error {
	output, err := exec.Command("fly", "-t", target, "builds", "-c", strconv.Itoa(count)).Output()
	if err != nil {
		return err
	}

	table := ui.Table{
		Headers: ui.TableRow{
			{Contents: "id", Color: color.New(color.Bold)},
			{Contents: "pipeline/job", Color: color.New(color.Bold)},
			{Contents: "build", Color: color.New(color.Bold)},
			{Contents: "status", Color: color.New(color.Bold)},
			{Contents: "start", Color: color.New(color.Bold)},
			{Contents: "end", Color: color.New(color.Bold)},
			{Contents: "duration", Color: color.New(color.Bold)},
		},
	}

	scanner := bufio.NewScanner(bytes.NewReader(output))
	for scanner.Scan() {
		cols := strings.Fields(scanner.Text())
		if len(cols) != len(table.Headers) {
			return errors.New("unable to parse line: " + scanner.Text())
		}

		cells := []ui.TableCell{
			ui.TableCell{Contents: cols[0]},
			ui.TableCell{Contents: cols[1]},
			ui.TableCell{Contents: cols[2]},
			ui.TableCell{Contents: cols[3], Color: colorize(cols[3])},
			ui.TableCell{Contents: cols[4]},
			ui.TableCell{Contents: cols[5]},
			ui.TableCell{Contents: cols[6]},
		}
		table.Data = append(table.Data, cells)
	}

	table.Render(os.Stdout)
	fmt.Fprint(os.Stdout, "\r\033[s\033["+strconv.Itoa(count+1)+"A")

	return nil
}

func colorize(status string) *color.Color {
	switch status {
	case "pending":
		return ui.PendingColor
	case "started":
		return ui.StartedColor
	case "succeeded":
		return ui.SucceededColor
	case "failed":
		return ui.FailedColor
	case "errored":
		return ui.ErroredColor
	case "aborted":
		return ui.AbortedColor
	case "paused":
		return ui.PausedColor
	}
	return nil
}
