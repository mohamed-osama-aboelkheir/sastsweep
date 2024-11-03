package runner

import (
	"errors"
	"net/http"
	"os"
	"os/signal"
	"sastsweep/common"
	"sastsweep/common/logger"
	"sastsweep/common/report"
	"sastsweep/common/sastsweep"
	"strconv"
	"strings"
	"sync"
	"syscall"

	"github.com/fatih/color"
)

func scanTarget(target string, options *common.Options, httpClient *http.Client) (sastsweep.RepoInfo, error) {

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT)
	go func() {
		sigCount := 0
		for sig := range sigChan {
			if sig != syscall.SIGINT {
				continue
			}

			sigCount++
			if sigCount == 2 {
				os.Exit(1)
			}
		}
	}()

	repoDoc, err := sastsweep.GetRepoDocument(httpClient, target)
	if errors.Is(err, common.Err404) {
		if options.Fireprox != "" {
			sastsweep.White.Printf("%s ", strings.Replace(target, options.Fireprox, "github.com/", 1))
		} else {
			sastsweep.White.Printf("%s ", target)
		}

		color.Red("[404]")
		return sastsweep.RepoInfo{}, err
	} else if err != nil {
		logger.Error("Could not create the repo doc for " + target + ": " + err.Error())
		return sastsweep.RepoInfo{}, err
	}

	repoInfo, err := sastsweep.ExtractRepoInfo(repoDoc, options)
	if err != nil {
		return repoInfo, err
	}
	repoInfo.Target = target

	if (!options.NoSemgrep || options.Branch) && repoInfo.Branch == "" {
		defaultBranch := sastsweep.GetDefaultBranch(repoDoc)
		repoInfo.Branch = defaultBranch
	}

	sourcePath := ""
	if !options.NoSemgrep || options.Files || options.FilterFiles != "" {
		sourcePath, err = sastsweep.DownloadSource(httpClient, target, repoInfo.Branch, options.OutDir)
		if err != nil && err != common.Err422 {
			logger.Error("Cannot download source code for " + target + ": " + err.Error())
			return repoInfo, err
		}

		if !options.SaveRepo {
			defer os.RemoveAll(sourcePath)
		}
	}

	if options.Files || options.FilterFiles != "" {
		numFiles, err := common.CountFiles(sourcePath)
		if err != nil {
			logger.Error("Could not count files in " + sourcePath + ": " + err.Error())
			return repoInfo, err
		}

		repoInfo.Files = numFiles

		if options.FilterFiles != "" {
			err := sastsweep.NumericFilter(repoInfo.Files, options.FilterFiles)
			if err != nil {
				return repoInfo, err
			}
		}
	}

	if !options.NoSemgrep {
		semgrepJson, err := sastsweep.RunSemgrep(sourcePath, options.OutDir)
		if err != nil {
			logger.Error("Problem running semgrep on " + sourcePath + ": " + err.Error())
			return repoInfo, err
		}

		repoInfo.SemgrepHits = len(semgrepJson.Results)

		var semgrepFindings []report.SemgrepFinding
		for _, result := range semgrepJson.Results {
			url := target + "/blob/" + repoInfo.Branch + "/" + strings.Join(strings.Split(result.Path, "/")[len(strings.Split(options.OutDir, "/"))+2:], "/") + "#L" + strconv.Itoa(result.Start.Line) + "-L" + strconv.Itoa(result.End.Line)
			if options.Github1s {
				url = strings.Replace(url, "github.com/", "github1s.com/", 1)
			}

			semgrepFindings = append(semgrepFindings, report.SemgrepFinding{
				VulnerabilityTitle: result.CheckID,
				Severity:           result.Extra.Metadata.Impact,
				Description:        result.Extra.Message,
				Code:               result.Extra.Lines,
				StartLine:          result.Start.Line,
				StopLine:           result.End.Line,
				GithubLink:         url,
			})
		}
		report.SortFindings(semgrepFindings)
		vulnerabilityStats, vulnerabilityStatsOrdering, severityStats := report.CalculateSemgrepMetrics(semgrepFindings)

		severityStatsOrdering := make([]string, 0, len(severityStats))
		for key := range severityStats {
			severityStatsOrdering = append(severityStatsOrdering, strings.ToUpper(key))
		}

		reportData := report.ReportData{
			Target:                     target,
			VulnerabilityStats:         vulnerabilityStats,
			VulnerabilityStatsOrdering: vulnerabilityStatsOrdering,
			SeverityStats:              severityStats,
			SeverityStatsOrdering:      severityStatsOrdering,
			Findings:                   semgrepFindings,
		}

		reportPath, err := report.GenerateHTML(reportData, options.OutDir)
		if err != nil {
			logger.Error("Unable to generate the HTML report for " + target + ": " + err.Error())
			return repoInfo, err
		}

		repoInfo.ReportPath = reportPath
	}

	return repoInfo, nil
}

func RepoScanner(targets <-chan string, options *common.Options, wg *sync.WaitGroup, stop chan bool) {
	defer wg.Done()

	httpClient := &http.Client{}

	for target := range targets {
		select {
		case <-stop:
			logger.Info("Stopping worker...")
			return
		default:

			if options.Fireprox != "" {
				target = strings.Replace(target, "github.com/", options.Fireprox, 1)
			}

			logger.Debug("Scanning " + target)
			repoInfo, err := scanTarget(target, options, httpClient)

			if err != nil && err != common.Err404 && err != common.Err422 && err != sastsweep.ErrTargetFiltered {
				logger.Error("Unable to scan " + target + ": " + err.Error())
			}

			if err == nil {
				sastsweep.DisplayRepoInfo(options, repoInfo)
			}
		}
	}
}
