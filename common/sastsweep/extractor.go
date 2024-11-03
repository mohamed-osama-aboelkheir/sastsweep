package sastsweep

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sastsweep/common"
	"sastsweep/common/logger"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

var errInvalidFilter = errors.New("invalid filter")
var ErrTargetFiltered = errors.New("could not pass filter")

func GetRepoDocument(client *http.Client, targetURL string) (*goquery.Document, error) {

	htmlBytes, err := common.HTTPGet(client, targetURL, nil)
	if err != nil {
		return nil, err
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(htmlBytes)))
	if err != nil {
		logger.Error("Could not create the goquery document: " + err.Error())
		return nil, err
	}

	return doc, nil
}

func ExtractRepoInfo(repoDoc *goquery.Document, options *common.Options) (RepoInfo, error) {
	title := repoDoc.Find("title").Text()
	parts := strings.Split(strings.Split(title, " ")[2], "/")
	repoOwner := parts[0]
	repoName := strings.Split(parts[1], ":")[0]
	repoPath := "/" + repoOwner + "/" + repoName
	target := "https://github.com" + repoPath

	repoInfo := RepoInfo{}

	if options.Description || options.FullDescription {
		fullDescription := strings.TrimSpace(repoDoc.Find("p.f4.my-3").Text())

		if options.FullDescription {
			repoInfo.Description = fullDescription
		} else if options.Description && len(fullDescription) > 80 {
			repoInfo.Description = fullDescription[:80]
		} else {
			repoInfo.Description = fullDescription
		}
	}

	if options.Stars || options.FilterStars != "" {
		stars := strings.TrimSpace(repoDoc.Find("a.Link.Link--muted[href='" + repoPath + "/stargazers']").Find("strong").Text())
		starCount, err := convertK(stars)
		if err != nil {
			logger.Error("Cannot convert " + target + " stars to kform " + stars)
			return repoInfo, err
		}
		repoInfo.Stars = starCount

		if options.FilterStars != "" {
			err := NumericFilter(repoInfo.Stars, options.FilterStars)
			if err != nil {
				return repoInfo, err
			}
		}
	}

	if options.Language {
		langDesc := ""
		repoDoc.Find("a.d-inline-flex.flex-items-center.flex-nowrap.Link--secondary.no-underline.text-small.mr-3:last-child").Each(func(index int, element *goquery.Selection) {
			langSplit := strings.Split(strings.TrimSpace(element.Text()), "\n")
			lang := strings.TrimSpace(langSplit[0])
			percentage := strings.TrimSpace(langSplit[1])

			langDesc += lang + ":" + percentage + ","
		})

		repoInfo.Language = langDesc
	}

	if options.Topics {
		topicDesc := ""
		repoDoc.Find("a.topic-tag-link").Each(func(index int, element *goquery.Selection) {
			topicDesc += strings.TrimSpace(element.Text()) + ","
		})

		repoInfo.Topics = topicDesc
	}

	if options.RepoLink {
		repoInfo.RepoLink = strings.TrimSpace(repoDoc.Find(".flex-auto.min-width-0.css-truncate.css-truncate-target.width-fit").First().Text())
	}

	if options.Forks || options.FilterForks != "" {
		forks := strings.TrimSpace(repoDoc.Find("a.Link.Link--muted[href='" + repoPath + "/forks']").Find("strong").Text())
		forkCount, err := convertK(forks)
		if err != nil {
			logger.Error("Cannot convert " + forks + " to kform: " + err.Error())
			return repoInfo, err
		}
		repoInfo.Forks = forkCount

		if options.FilterForks != "" {
			err := NumericFilter(repoInfo.Forks, options.FilterForks)
			if err != nil {
				return repoInfo, err
			}
		}
	}

	if options.Branch {
		repoInfo.Branch = GetDefaultBranch(repoDoc)
	}

	if options.Commits || options.FilterCommits != "" {
		repoInfo.Commits = getCommitCount(repoDoc)

		if options.FilterCommits != "" {
			err := NumericFilter(repoInfo.Commits, options.FilterCommits)
			if err != nil {
				return repoInfo, err
			}
		}
	}

	if options.LastRelease || options.FilterLastRelease != "" {
		date := strings.TrimSpace(repoDoc.Find("relative-time").First().Text())
		if date != "" {
			lastRelease, err := time.Parse("Jan 2, 2006", date)
			if err != nil {
				logger.Error("Could not convert date " + date + " to datetime: " + err.Error())
				return repoInfo, err
			}
			repoInfo.LastRelease = lastRelease
		}

		if date != "" && !repoInfo.LastRelease.IsZero() && options.FilterLastRelease != "" {
			err := dateFilter(repoInfo.LastRelease, options.FilterLastRelease)
			if err != nil {
				return repoInfo, err
			}
		}
	}

	if options.LastCommit || options.FilterLastCommit != "" {
		if repoInfo.Branch == "" {
			repoInfo.Branch = GetDefaultBranch(repoDoc)
		}
		lastCommit, err := getLastCommit(target, repoInfo.Branch, repoDoc)
		if err != nil {
			logger.Error("Could not obtain the last commit date: " + err.Error())
			return repoInfo, err
		}

		if !repoInfo.LastCommit.IsZero() {
			repoInfo.LastCommit = lastCommit
		}

		if !repoInfo.LastCommit.IsZero() && options.FilterLastCommit != "" {
			err := dateFilter(repoInfo.LastCommit, options.FilterLastCommit)
			if err != nil {
				return repoInfo, err
			}
		}
	}

	if options.FirstCommit || options.FilterFirstCommit != "" {
		if repoInfo.Branch == "" {
			repoInfo.Branch = GetDefaultBranch(repoDoc)
		}
		if repoInfo.Commits == 0 {
			repoInfo.Commits = getCommitCount(repoDoc)
		}
		firstCommit := getFirstCommit(target, repoInfo.Commits, repoInfo.Branch, repoDoc)
		if !firstCommit.IsZero() {
			repoInfo.FirstCommit = firstCommit
		}

		if !repoInfo.FirstCommit.IsZero() && options.FilterFirstCommit != "" {
			err := dateFilter(repoInfo.FirstCommit, options.FilterFirstCommit)
			if err != nil {
				return repoInfo, err
			}
		}
	}

	if options.SecurityIssues || options.FilterSecurityIssues != "" {
		repoInfo.SecurityIssues = getSecurityIssues(target)

		if options.FilterSecurityIssues != "" {
			err := NumericFilter(repoInfo.SecurityIssues, options.FilterSecurityIssues)
			if err != nil {
				return repoInfo, err
			}
		}
	}

	if options.Watchers || options.FilterWatchers != "" {
		watchers := strings.TrimSpace(repoDoc.Find("a.Link.Link--muted[href='" + repoPath + "/watchers']").Find("strong").Text())
		watcherCount, err := strconv.Atoi(watchers)
		if err != nil {
			logger.Error("Could not convert watchers " + watchers + " to int: " + err.Error())
			return repoInfo, err
		}
		repoInfo.Watchers = watcherCount

		if options.FilterWatchers != "" {
			err := NumericFilter(repoInfo.Watchers, options.FilterWatchers)
			if err != nil {
				return repoInfo, err
			}
		}
	}

	if options.Contributors || options.FilterContributors != "" {
		contributors := strings.Replace(strings.Replace(strings.Replace(strings.TrimSpace(repoDoc.Find("a.Link--inTextBlock.Link[href='"+repoPath+"/graphs/contributors']").Text()), ",", "", -1), " contributors", "", -1), "+ ", "", -1)
		if contributors == "" {
			repoInfo.Contributors = 0
		} else {
			contributorCount, err := strconv.Atoi(contributors)
			if err != nil {
				logger.Error("Could not convert contributors " + contributors + " to int: " + err.Error())
				return repoInfo, err
			}
			repoInfo.Contributors = contributorCount
		}

		if options.FilterContributors != "" {
			err := NumericFilter(repoInfo.Contributors, options.FilterContributors)
			if err != nil {
				return repoInfo, err
			}
		}
	}

	if options.Issues || options.FilterIssues != "" {
		issues := strings.Replace(strings.TrimSpace(repoDoc.Find(".Counter#issues-repo-tab-count").Text()), ",", "", -1)
		issueCount, err := strconv.Atoi(issues)
		if err != nil {
			logger.Error("Could not convert issues " + issues + " to int: " + err.Error())
			return repoInfo, err
		}
		repoInfo.Issues = issueCount

		if options.FilterIssues != "" {
			err := NumericFilter(repoInfo.Issues, options.FilterIssues)
			if err != nil {
				return repoInfo, err
			}
		}
	}

	if options.PullRequests || options.FilterPullRequests != "" {
		pullRequestCount := strings.Replace(strings.TrimSpace(repoDoc.Find(".Counter#pull-requests-repo-tab-count").Text()), ",", "", -1)
		pullRequests, err := strconv.Atoi(pullRequestCount)
		if err != nil {
			logger.Error("Could not convert pull request count " + pullRequestCount + " to int: " + err.Error())
			return repoInfo, err
		}
		repoInfo.PullRequests = pullRequests

		if options.FilterPullRequests != "" {
			err := NumericFilter(repoInfo.PullRequests, options.FilterPullRequests)
			if err != nil {
				return repoInfo, err
			}
		}
	}

	return repoInfo, nil
}

func NumericFilter(value int, filter string) error {
	var from int
	var to int
	var err error

	filterSplit := strings.Split(filter, "-")
	if len(filterSplit) == 1 {
		from, err = strconv.Atoi(filterSplit[0])
		if err != nil {
			logger.Error("Unable to convert filter string " + filterSplit[0] + " to int " + err.Error())
		}
		to = from
	} else if len(filterSplit) == 2 && filterSplit[0] != "" && filterSplit[1] != "" {
		from, err = strconv.Atoi(filterSplit[0])
		if err != nil {
			logger.Error("Unable to convert filter string " + filterSplit[0] + " to int " + err.Error())
			return errInvalidFilter
		}

		to, err = strconv.Atoi(filterSplit[1])
		if err != nil {
			logger.Error("Unable to convert filter string " + filterSplit[1] + " to int " + err.Error())
			return errInvalidFilter
		}
	} else if len(filterSplit) == 2 && filterSplit[0] != "" && filterSplit[1] == "" {
		from, err = strconv.Atoi(filterSplit[0])
		if err != nil {
			logger.Error("Unable to convert filter string " + filterSplit[0] + " to int " + err.Error())
			return errInvalidFilter
		}

		to = -1
	} else if len(filterSplit) == 2 && filterSplit[0] == "" && filterSplit[1] != "" {
		from = 0

		to, err = strconv.Atoi(filterSplit[1])
		if err != nil {
			logger.Error("Unable to convert filter string " + filterSplit[0] + " to int " + err.Error())
			return errInvalidFilter
		}
	} else {
		logger.Error("The filterstring " + filter + " is invalid")
		return errInvalidFilter
	}

	if (value >= from && value <= to) || (value >= from && to == -1) {
		return nil
	}

	return ErrTargetFiltered
}

func convertK(s string) (int, error) {
	if !strings.Contains(s, "k") {
		sint, err := strconv.Atoi(s)
		if err != nil {
			logger.Error("Could not convert " + s + " to int")
			return 0, err
		}
		return sint, nil
	}

	s = strings.TrimSpace(strings.ToLower(s))

	if !strings.HasSuffix(s, "k") {
		num, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return 0, err
		}
		return int(num), nil
	}

	numStr := strings.TrimSuffix(s, "k")

	num, err := strconv.ParseFloat(numStr, 64)
	if err != nil {
		return 0, err
	}

	return int(num * 1000), nil
}

func GetDefaultBranch(repoDoc *goquery.Document) string {
	jsonContent := repoDoc.Find("script[type='application/json'][data-target='react-partial.embeddedData']").Last().Text()

	branchIndex := strings.Index(jsonContent, `"defaultBranch":`)
	startIndex := branchIndex + len(`"defaultBranch":"`)
	endIndex := strings.Index(jsonContent[startIndex:], `"`) + startIndex
	branchValue := jsonContent[startIndex:endIndex]

	return branchValue
}

func getSecurityIssues(target string) int {
	respBytes, err := common.HTTPGet(&http.Client{}, target+"/security/overall-count", map[string]string{"Accept": "*/*"})
	if err != nil {
		return 0
	}

	securityDoc, err := goquery.NewDocumentFromReader(strings.NewReader(string(respBytes)))
	if err != nil {
		return 0
	}

	securityIssues, err := strconv.Atoi(strings.TrimSpace(securityDoc.Find("span").Text()))
	if err != nil {
		return 0
	}

	return securityIssues
}

func getCommitCount(repoDoc *goquery.Document) int {
	commits := strings.Replace(strings.Replace(strings.TrimSpace(repoDoc.Find(".fgColor-default").First().Text()), ",", "", -1), " Commits", "", -1)
	commitCount, err := strconv.Atoi(commits)
	if err != nil {
		logger.Error("Cannot convert commit count " + commits + " to int: " + err.Error())
	} else {
		return commitCount
	}

	return 0
}

func getFirstCommit(target string, commits int, branch string, repoDoc *goquery.Document) time.Time {
	jsonContent := repoDoc.Find("script[type='application/json'][data-target='react-partial.embeddedData']").Last().Text()
	oidIndex := strings.Index(jsonContent, `"currentOid":`)
	startIndex := oidIndex + 14 // Length of `"currentOid":"`
	endIndex := startIndex + 40 // SHA-1 hash is 40 characters long
	hashValue := jsonContent[startIndex:endIndex]

	url := fmt.Sprintf("%s/commits/%s?after=%s+%d", target, branch, hashValue, commits-2)
	respBytes, err := common.HTTPGet(&http.Client{}, url, nil)
	if err != nil {
		return time.Time{}
	}

	oldestCommitDoc, err := goquery.NewDocumentFromReader(strings.NewReader(string(respBytes)))
	if err != nil {
		return time.Time{}
	}

	dateStr := strings.TrimPrefix(strings.TrimSpace(oldestCommitDoc.Find("h3.text-normal").First().Text()), "Commits on ")
	oldestCommitDate, err := time.Parse("Jan 2, 2006", dateStr)
	if err != nil {
		return time.Time{}
	}

	return oldestCommitDate

}

func getLastCommit(target string, branch string, repoDoc *goquery.Document) (time.Time, error) {

	respBytes, err := common.HTTPGet(&http.Client{}, target+"/latest-commit/"+branch, map[string]string{"Accept": "application/json"})
	if err != nil {
		return time.Time{}, err
	}

	var commitInfo CommitInfo
	err = json.Unmarshal(respBytes, &commitInfo)
	if err != nil {
		return time.Time{}, err
	}

	t, err := time.Parse(time.RFC3339, commitInfo.Date)
	if err != nil {
		logger.Error("Unable to parse " + commitInfo.Date + " into date format: " + err.Error())
		return time.Time{}, err
	}

	lastCommit := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())

	return lastCommit, nil
}

func dateFilter(targetDate time.Time, filter string) error {
	var fromDate time.Time
	var toDate time.Time
	var err error

	filterSplit := strings.Split(strings.Replace(filter, "/", "", -1), "-")
	if len(filterSplit) == 1 {
		fromDate, err = time.Parse("20060102", filterSplit[0])
		if err != nil {
			logger.Error("Unable to parse time from " + filterSplit[0] + ": " + err.Error())
		}

		toDate = fromDate
	} else if len(filterSplit) == 2 && filterSplit[0] != "" && filterSplit[1] != "" {
		fromDate, err = time.Parse("20060102", filterSplit[0])
		if err != nil {
			logger.Error("Unable to parse time from " + filterSplit[0] + ": " + err.Error())
		}

		toDate, err = time.Parse("20060102", filterSplit[1])
		if err != nil {
			logger.Error("Unable to parse time from " + filterSplit[1] + ": " + err.Error())
		}
	} else if len(filterSplit) == 2 && filterSplit[0] != "" && filterSplit[1] == "" {
		fromDate, err = time.Parse("20060102", filterSplit[0])
		if err != nil {
			logger.Error("Unable to parse time from " + filterSplit[0] + ": " + err.Error())
		}

		toDate = time.Time{}
	} else if len(filterSplit) == 2 && filterSplit[0] == "" && filterSplit[1] != "" {
		fromDate = time.Time{}

		toDate, err = time.Parse("20060102", filterSplit[1])
		if err != nil {
			logger.Error("Unable to parse time from " + filterSplit[1] + ": " + err.Error())
		}
	} else {
		logger.Error("The filterstring " + filter + " is invalid")
		return errInvalidFilter
	}

	if fromDate.IsZero() {
		if targetDate.Before(toDate) || targetDate.Equal(toDate) {
			return nil
		}
	} else if toDate.IsZero() {
		if targetDate.After(fromDate) || targetDate.Equal(fromDate) {
			return nil
		}
	} else {
		if (targetDate.Before(toDate) || targetDate.Equal(toDate)) && (targetDate.After(fromDate) || targetDate.Equal(fromDate)) {
			return nil
		}
	}

	return ErrTargetFiltered
}
