package sastsweep

import (
	"bufio"
	"encoding/json"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/chebuya/sastsweep/common"
	"github.com/chebuya/sastsweep/common/logger"

	"github.com/google/uuid"
)

var localBinPath = ""
var semgrepBinaryPath = ""
var semgrepConfig = []string{}

func ConfigureSemgrep(outDir string, configPath string, semgrepPath string) error {

	dirname, err := os.UserHomeDir()
	if err != nil {
		logger.Error("Unable to determine the user home directory: " + err.Error())
		return err
	}

	localBinPath = filepath.Join(dirname, ".local", "bin")
	if semgrepPath == "" {
		semgrepBinaryPath = filepath.Clean(filepath.Join(localBinPath, "semgrep"))
	}
	if semgrepPath != "" {
		semgrepBinaryPath = semgrepPath
	}

	if configPath == "" {
		configPath = filepath.Clean(filepath.Join(strings.TrimRight(outDir, "/scans"), "sastsweep.conf"))
	}

	_, err = os.Stat(semgrepBinaryPath)
	if os.IsNotExist(err) {
		logger.Info("Semgrep not found, installing...")
		cmd := exec.Command("pip3", "install", "semgrep", "--break-system-packages")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		retry := false
		if err := cmd.Run(); err != nil {
			logger.Error("Could not install semgrep, trying again: " + err.Error())
			retry = true
		}

		if retry {
			cmd = exec.Command("pip", "install", "semgrep", "--break-system-packages")
			if err := cmd.Run(); err != nil {
				logger.Error("Second semgrep install attempt failed: " + err.Error())
				return err
			}
		}
	}

	_, err = os.Stat(configPath)
	if os.IsNotExist(err) {
		semgrepConfig = []string{"--config", "auto"}
		return nil
	} else if err != nil {
		logger.Error("Problem configuring semgrep flags: " + err.Error())
		return err
	}

	file, err := os.Open(configPath)
	if err != nil {
		logger.Error("Could not open semgrep config file: " + err.Error())
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "FLAGS=") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		semgrepConfig = strings.Split(parts[1], " ")
		break
	}

	return nil
}

func RunSemgrep(sourcePath string, outDir string) (SemgrepJson, error) {
	var semgrepJson SemgrepJson

	semgrepOutFile := outDir + "/output-" + uuid.New().String() + ".json"

	args := []string{"--json", "--config", "p/python", "--output", semgrepOutFile}
	args = append(args, semgrepConfig...)
	args = append(args, sourcePath)

	cmd := exec.Command(semgrepBinaryPath, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
		Pgid:    0,
	}
	path := os.Getenv("PATH") + localBinPath
	cmd.Env = append(os.Environ(), "PATH="+path+":"+localBinPath)

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
				if err := cmd.Process.Kill(); err != nil {
					logger.Error("Failed to force killing of semgrep: " + err.Error())
				}
				return
			}
		}
	}()

	if err := cmd.Start(); err != nil {
		logger.Error("Could not start semgrep on " + sourcePath + ": " + err.Error())
		return semgrepJson, err
	}

	err := cmd.Wait()
	if err != nil {
		logger.Error("semgrep finished with error: " + err.Error())
		return semgrepJson, err
	}

	defer os.Remove(semgrepOutFile)

	fileBytes, err := common.ReadFile(semgrepOutFile)
	if err != nil {
		logger.Error("Could not read from " + semgrepOutFile + ": " + err.Error())
		return semgrepJson, err
	}

	if err := json.Unmarshal(fileBytes, &semgrepJson); err != nil {
		logger.Error("Could not unmarshal " + sourcePath + " semgrep output: " + err.Error())
		return semgrepJson, err
	}

	return semgrepJson, nil
}
