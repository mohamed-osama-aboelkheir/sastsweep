package report

import (
	"html/template"
	"os"

	"github.com/chebuya/sastsweep/common/logger"

	"github.com/google/uuid"
)

func GenerateHTML(reportData ReportData, outDir string) (string, error) {
	tmpl := template.New("report")
	tmpl.Funcs(template.FuncMap{
		"getLanguage": getLanguage,
		"toLowerCase": toLowerCase,
	})

	// Parse the template
	tmpl, err := tmpl.Parse(htmlTemplate)
	if err != nil {
		logger.Error("Could not parse the template: " + err.Error())
		return "", err
	}

	// Create output file
	outPath := outDir + "/report-" + uuid.New().String() + ".html"
	file, err := os.Create(outPath)
	if err != nil {
		logger.Error("Could not create the output HTML file: " + err.Error())
		return "", err
	}
	defer file.Close()

	// Execute template with semgrepFindings
	err = tmpl.Execute(file, reportData)
	if err != nil {
		logger.Error("Could not execute the HTML template: " + err.Error())
		return "", err
	}

	return outPath, nil
}
