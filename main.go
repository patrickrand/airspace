package main

import (
	"bufio"
	"bytes"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"time"

	"github.com/concourse/fly/ui"
	"github.com/fatih/color"
)

var (
	countFlag       = flag.Int("c", 10, "count")
	targetFlag      = flag.String("t", "local", "target")
	pipelineJobFlag = flag.String("p", "", "pipeline/job regex")

	table = ui.Table{
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
)

func main() {
	flag.Parse()
	table.Data = make([]ui.TableRow, *countFlag)

	disableInputBuffering()
	disableStdinDisplay()

	kill := make(chan struct{})
	go exitHandler(kill)

	for {
		select {
		case <-time.After(1 * time.Second):
			if err := run(*pipelineJobFlag, *countFlag); err != nil {
				panic(err)
			}
		case <-kill:
			terminate()
		}
	}
}

func run(pattern string, count int) error {
	output, err := exec.Command("fly", "-t", *targetFlag, "builds").Output()
	if err != nil {
		return err
	}

	var i int
	scanner := bufio.NewScanner(bytes.NewReader(output))
	for scanner.Scan() && i < count {
		cols := strings.Fields(scanner.Text())
		if len(cols) != len(table.Headers) {
			return errors.New("unable to parse line: " + scanner.Text())
		}

		if pattern != "" { // TODO: implement pattern matching
			return errors.New("-pj (pipeline/job) is currently not implemented")
		}

		table.Data[i] = []ui.TableCell{
			ui.TableCell{Contents: cols[0]},
			ui.TableCell{Contents: cols[1]},
			ui.TableCell{Contents: cols[2]},
			ui.TableCell{Contents: cols[3], Color: colorize(cols[3])},
			ui.TableCell{Contents: cols[4]},
			ui.TableCell{Contents: cols[5]},
			ui.TableCell{Contents: cols[6]},
		}
		i++
	}

	table.Render(os.Stdout)
	fmt.Fprintf(os.Stdout, "\r\033[?25l\033[s\033[%dA", i+1) // move ANSI cursor to upper-left corner
	return nil
}

func disableInputBuffering() {
	if err := exec.Command("stty", "-f", "/dev/tty", "cbreak", "min", "1").Run(); err != nil {
		if err := exec.Command("stty", "-F", "/dev/tty", "cbreak", "min", "1").Run(); err != nil {
			panic(err)
		}
	}
}

func disableStdinDisplay() {
	if err := exec.Command("stty", "-f", "/dev/tty", "-echo").Run(); err != nil {
		if err := exec.Command("stty", "-F", "/dev/tty", "-echo").Run(); err != nil {
			panic(err)
		}
	}
}

func exitHandler(kill chan struct{}) {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)
	go func() {
		for range signals {
			kill <- struct{}{} // TODO: properly handle each signal type
		}
	}()

	b := make([]byte, 1)
	for {
		os.Stdin.Read(b)
		if b[0] == 'q' {
			kill <- struct{}{}
		}
	}
}

func terminate() {
	table.Render(os.Stdout)            // dump latest table to stdout
	fmt.Fprint(os.Stdout, "\033[?25h") // re-enable ANSI cursor
	os.Exit(0)
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
