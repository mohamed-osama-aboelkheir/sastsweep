package sastsweep

import (
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/chebuya/sastsweep/common"
	"github.com/chebuya/sastsweep/common/logger"
)

func DownloadSource(client *http.Client, targetURL string, branch string, outDir string) (string, error) {
	target := strings.Replace(strings.Replace(targetURL, "http://", "", 1), "https://", "", 1)
	outPath := filepath.Clean(filepath.Join(outDir, strings.Split(target, "/")[1]+"-"+strings.Split(target, "/")[2]))

	if !strings.HasPrefix(outPath, filepath.Clean(outDir)+string(os.PathSeparator)) {
		logger.Debug(outPath + " is outside of output directory: " + outDir)
		return "", errors.New("download path outside of output directory")
	}

	if _, err := os.Stat(outPath); err == nil {
		os.RemoveAll(outPath)
		os.Exit(1)
	}

	err := os.Mkdir(outPath, 0700)
	if err != nil {
		logger.Error("There was an error creating the output directory " + outPath + ": " + err.Error())
		return "", err
	}

	zipBytes, err := common.HTTPGet(client, targetURL+"/archive/refs/heads/"+branch+".zip", nil)
	if err != nil {
		return "", err
	}

	err = common.UnzipBytes(zipBytes, outPath)
	if err != nil {
		logger.Error("Could not unzip zipBytes to " + outPath + ": " + err.Error())
	}

	return outPath, nil
}
