package main

import (
	"fmt"
	"os/exec"
	"strconv"
	"time"
)

func main() {
	for range time.Tick(2 * time.Second) {
		printBuilds("mia", "", 20)
	}
}

func printBuilds(target, pattern string, count int) error {
	output, err := exec.Command("fly", "-t", target, "builds", "-c", strconv.Itoa(count)).CombinedOutput()
	if err != nil {
		return err
	}

	fmt.Printf("\r%s", string(output))
	fmt.Printf("\033[s\033[" + strconv.Itoa(count) + "A")
	return nil
}
