package report

const htmlTemplate = `
<!DOCTYPE html>
<html lang="en">
<title>Semgrep Findings</title>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/prism/1.24.1/themes/prism-tomorrow.min.css">
    <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/prism/1.24.1/plugins/line-numbers/prism-line-numbers.min.css">
    <style>
        :root {
            --cyan: #00fff2;
            --magenta: #ff00ff;
            --dark-bg: #0a0a0f;
            --card-bg: #151520;
            --text-color: #e0e0e0;
            --gradient-cyan: rgba(0, 255, 242, 0.05);
            --gradient-magenta: rgba(255, 0, 255, 0.05);
        }

        body {
            font-family: 'Courier New', monospace;
            line-height: 1.6;
            margin: 0;
            padding: 40px 20px;
            background: var(--dark-bg);
            color: var(--text-color);
            background-image: 
                linear-gradient(45deg, var(--gradient-cyan) 1px, transparent 1px),
                linear-gradient(-45deg, var(--gradient-magenta) 1px, transparent 1px);
            background-size: 50px 50px;
            min-height: 100vh;
        }

        a {
            color: var(--cyan);
            text-decoration: none;
            transition: all 0.3s ease-in-out;
            border-bottom: 1px solid rgba(0, 255, 242, 0.3);
            padding-bottom: 2px;
            position: relative;
        }

        a:hover {
            color: var(--magenta);
            border-bottom-color: var(--magenta);
            text-shadow: 0 0 8px var(--magenta);
            padding-left: 5px;
        }

        a:visited {
            color: var(--cyan);
            border-bottom-color: rgba(0, 255, 242, 0.3);
        }

        .vulnerability-card {
            background: var(--card-bg);
            border-radius: 12px;
            box-shadow: 
                0 0 20px rgba(0, 255, 242, 0.1),
                0 0 40px rgba(0, 255, 242, 0.05);
            margin-bottom: 30px;
            padding: 25px;
            border: 1px solid rgba(0, 255, 242, 0.2);
            position: relative;
            overflow: hidden;
            transition: transform 0.3s ease, box-shadow 0.3s ease;
        }

        .filepath {
            font-family: 'Courier New', monospace;
            background: rgba(0, 255, 242, 0.08);
            padding: 8px 12px;
            border-radius: 6px;
            color: var(--magenta);
            font-size: 0.9em;
            border-left: 3px solid var(--magenta);
            margin-bottom: 15px;
            word-break: break-all;
        }

        .vulnerability-name {
            color: var(--magenta);
            font-weight: bold;
            margin: 15px 0;
            font-size: 1.3em;
            text-transform: uppercase;
            letter-spacing: 1px;
            text-shadow: 0 0 10px rgba(255, 0, 255, 0.3);
        }

        .severity {
            display: inline-block;
            padding: 6px 14px;
            border-radius: 6px;
            font-size: 0.8em;
            font-weight: bold;
            text-transform: uppercase;
            letter-spacing: 1px;
            margin-left: 10px;
            box-shadow: 0 2px 4px rgba(0, 0, 0, 0.2);
        }

        .HIGH {
            background: rgba(255, 0, 0, 0.15);
            color: #ff4444;
            border: 1px solid rgba(255, 68, 68, 0.3);
        }

        .MEDIUM {
            background: rgba(255, 166, 0, 0.15);
            color: #ffaa00;
            border: 1px solid rgba(255, 170, 0, 0.3);
        }

        .LOW {
            background: rgba(0, 255, 0, 0.15);
            color: #00ff00;
            border: 1px solid rgba(0, 255, 0, 0.3);
        }

        .statistics-section {
            background: var(--card-bg);
            padding: 25px;
            border-radius: 12px;
            margin-bottom: 40px;
            color: #a0a0a0;
            box-shadow: 
                0 0 20px rgba(0, 255, 242, 0.1),
                0 0 40px rgba(0, 255, 242, 0.05);
            border: 1px solid rgba(0, 255, 242, 0.2);
        }

        .stat-item {
            margin: 15px 0;
            transition: background-color 0.3s ease;
        }


        .stat-item:hover {
            background: rgba(0, 255, 242, 0.05);
        }

        .vulnerability-card:hover {
            transform: translateY(-2px);
            box-shadow: 
                0 0 25px rgba(0, 255, 242, 0.15),
                0 0 50px rgba(0, 255, 242, 0.1);
        }

        .severity-stats {
            display: flex;
            flex-wrap: wrap;
            gap: 15px;
            font-size: 0.9em;
            margin-bottom: 20px;
        }

        .vuln-name-stats {
            flex-wrap: wrap;
            font-size: x-small;
            text-align: left;
        }


        .target-section {
            text-align: left;
            color: var(--magenta);
            font-weight: bold;
            background: rgba(0, 255, 242, 0.08);
            border-radius: 12px;
            padding: 20px;
            margin-bottom: 30px;
            font-size: 1.1em;
            border: 1px solid rgba(0, 255, 242, 0.2);
            letter-spacing: 1px;
            box-shadow: 
                0 0 20px rgba(0, 255, 242, 0.1),
                0 0 40px rgba(0, 255, 242, 0.05);
        }

        .code-container {
            margin-top: 20px;
            border-radius: 8px;
            overflow: hidden;
        }

        pre {
            margin: 0;
            border-radius: 8px;
            background: rgba(0, 0, 0, 0.3) !important;
        }

        .description {
            margin: 15px 0;
            line-height: 1.7;
            color: #b0b0b0;
            padding: 10px;
            border-radius: 6px;
            background: rgba(0, 255, 242, 0.03);
        }

        @media (max-width: 600px) {
            .severity-stats {
                flex-direction: column;
            }
            
            .vuln-name-stats {
                flex-direction: column;
                align-items: center;
            }
        }

    </style>
</head>
<body>
	<div class="target-section">
        Target: <a href={{ .Target }}>{{ .Target }}</a>
    </div>

    <div class="statistics-section">
        <div class="severity-stats">
			{{ range .SeverityStatsOrdering }}
            <div class="stat-item">
                <span class="severity {{ . }}">{{ . }}</span>: {{ index $.SeverityStats . }} occurrences
            </div>
			{{ end }}
        </div>

        <div class="vuln-name-stats">
			{{ range .VulnerabilityStatsOrdering }}
            <div class="stat-item">
                <span class="vulnerability-name">{{ . }}</span>: {{ index $.VulnerabilityStats . }} occurrences
            </div>
			{{ end }}
        </div>
    </div>

    {{range .Findings}}
    <div class="vulnerability-card">
        <div class="filepath"><a href={{ .GithubLink }}>{{ .GithubLink }}</a></div>
        <div class="vulnerability-name">{{.VulnerabilityTitle}}</div>
        <div class="severity {{.Severity}}">{{.Severity}}</div>
        <div class="description">
            {{.Description}}
        </div>
        <div class="code-container">
            <pre class="line-numbers" data-start="{{.StartLine}}"><code class="language-{{getLanguage .VulnerabilityTitle}}">{{.Code}}</code></pre>
        </div>
    </div>
    {{end}}

    <script src="https://cdnjs.cloudflare.com/ajax/libs/prism/1.24.1/prism.min.js"></script>
    <script src="https://cdnjs.cloudflare.com/ajax/libs/prism/1.24.1/components/prism-javascript.min.js"></script>
    <script src="https://cdnjs.cloudflare.com/ajax/libs/prism/1.24.1/components/prism-python.min.js"></script>
    <script src="https://cdnjs.cloudflare.com/ajax/libs/prism/1.24.1/components/prism-jsx.min.js"></script>
    <script src="https://cdnjs.cloudflare.com/ajax/libs/prism/1.24.1/plugins/line-numbers/prism-line-numbers.min.js"></script>
</body>
</html>`
