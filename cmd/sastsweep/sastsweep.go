package main

import (
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"github.com/chebuya/sastsweep/common/feed"
	"github.com/chebuya/sastsweep/common/logger"
	"github.com/chebuya/sastsweep/common/sastsweep"
	"github.com/chebuya/sastsweep/runner"
)

func main() {
	options := runner.ParseOptions()

	if !options.NoSemgrep {
		err := sastsweep.ConfigureSemgrep(options.OutDir, options.ConfigPath, options.SemgrepPath)
		if err != nil {
			logger.Error("Could not configure semgrep: " + err.Error())
		}
	}

	runner.ShowBanner(options.NoEmoji)
	if options.NoSemgrep && options.Fireprox == "" {
		logger.Info("You are running SASTSweep without semgrep, you may want to use the -fireprox flag to avoid rate limiting")
	}

	if _, err := os.Stat(options.OutDir); os.IsNotExist(err) {
		err = os.MkdirAll(options.OutDir, 0700)
		if err != nil {
			logger.Error("Could not create " + options.OutDir + " as a directory: " + err.Error())
		}
	}

	targets := make(chan string, options.Threads*20)
	stop := make(chan bool)
	kill := make(chan os.Signal, 1)
	signal.Notify(kill, os.Interrupt)
	go func() {
		sigCount := 0
		for sig := range kill {
			if sig != syscall.SIGINT {
				continue
			}

			sigCount++

			if sigCount == 1 {
				logger.Info("Received CTRL+C, shutting down gracefully")
				logger.Info("Press CTRL+C again to force exit")
				for i := 0; i < options.Threads+1; i++ {
					stop <- true
				}
			} else if sigCount == 2 {
				os.Exit(1)
			}
		}
	}()

	var wg sync.WaitGroup
	for i := 0; i < options.Threads; i++ {
		wg.Add(1)
		go runner.RepoScanner(targets, options, &wg, stop)
	}

	if options.Repo != "" {
		logger.Info("Running on a single target")
		target := options.Repo
		if !strings.Contains(target, "github.com/") {
			logger.Error(target + " is not a valid github.com url")
		}

		target = strings.Replace(target, "http://", "https://", 1)
		if !strings.Contains(target, "https://") {
			target = "https://" + target
		}

		targets <- target
		close(targets)
	} else if options.Repos != "" {
		logger.Info("Running from a list of targets")
		go feed.FromFile(options.Repos, targets, stop)
	} else {
		logger.Info("Running using stdin")
		go feed.FromStdIn(targets, stop)
	}

	wg.Wait()
	if options.NoEmoji {
		logger.Info("All targets have been scanned.  Goodbye <3")
	} else {
		logger.Info("All targets have been scanned.  Goodbye 🩵")
	}
}
