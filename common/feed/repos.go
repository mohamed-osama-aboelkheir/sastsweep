package feed

import (
	"bufio"
	"os"
	"regexp"
	"slices"
	"strings"

	"github.com/chebuya/sastsweep/common/logger"
)

const githubUrlRegex = `\b^(https?://)?github.com/(?:[A-Za-z0-9-]+\.?)*[A-Za-z0-9-]+/(?:[A-Za-z0-9-]+\.?)*[A-Za-z0-9-]+$\b`

func fromScanner(scanner *bufio.Scanner, targets chan<- string, stop chan bool) {
	defer close(targets)

	alreadyScanned := []string{}
	targetRegex := regexp.MustCompile(githubUrlRegex)

	for scanner.Scan() {
		select {
		case <-stop:
			logger.Info("Stopping feeder...")
			return
		default:
			target := scanner.Text()
			if !targetRegex.MatchString(target) {
				continue
			}

			target = strings.Replace(target, "http://", "https://", 1)

			if !strings.Contains(target, "https://") {
				target = "https://" + target
			}

			if slices.Contains(alreadyScanned, target) {
				continue
			}
			alreadyScanned = append(alreadyScanned, target)
			targets <- target

		}
	}

	if err := scanner.Err(); err != nil {
		logger.Error("The input scanner encountered an error: " + err.Error())
	}

	logger.Debug("Finished reading stdin")
}

func FromStdIn(targets chan string, stop chan bool) {
	scanner := bufio.NewScanner(os.Stdin)
	fromScanner(scanner, targets, stop)
}

func FromFile(filename string, targets chan string, stop chan bool) {
	file, err := os.Open(filename)
	if err != nil {
		logger.Error("Could not read the input file " + filename + ": " + err.Error())
		os.Exit(1)
		return
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	fromScanner(scanner, targets, stop)
}
