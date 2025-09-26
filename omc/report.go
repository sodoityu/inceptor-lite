package omc

import (
	"fmt"
	"html"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

var (
	reportDir  = "."
	htmlOutput strings.Builder
	htmlOnce   sync.Once
	sectionIDs []string
)

func SetReportDir(dir string) {
	if dir != "" {
		reportDir = dir
	}
}
func WriteStdoutAndHTML(line string) {
	fmt.Println(line)
	writeHTMLLine(html.EscapeString(line))
}
func writeHTMLLine(line string) {
	htmlOutput.WriteString(line + "<br>\n")
}
func AppendSection(title, body string) {
	// 生成 id 供 TOC 跳转
	sectionID := strings.ToLower(title)
	sectionID = strings.ReplaceAll(sectionID, " ", "-")
	sectionID = strings.ReplaceAll(sectionID, "/", "-")
	sectionID = strings.ReplaceAll(sectionID, ":", "-")
	// 只把顶层标题放入 TOC
	if strings.HasPrefix(title, "Operator") || strings.HasPrefix(title, "Global") ||
		title == "Must-Gather Loaded" || title == "Node Status" {
		sectionIDs = append(sectionIDs, sectionID+"|"+title)
	}
	// 每个 Operator 开头加分隔符
	if strings.HasPrefix(title, "Operator") && !strings.Contains(title, "Logs") {
		htmlOutput.WriteString("<hr style='border:1px solid #ddd; margin:20px 0;'>\n")
	}
	// 如果标题包含 Degraded，加红色 :x:
	var displayTitle string
	if strings.Contains(strings.ToLower(title), "degraded") {
		displayTitle = fmt.Sprintf("<span style='color:red;'>&#10060; %s</span>", html.EscapeString(title))
	} else {
		displayTitle = html.EscapeString(title)
	}
	// 生成 HTML Section
	htmlOutput.WriteString(fmt.Sprintf("<details id=\"%s\"><summary><b>%s</b></summary>\n", sectionID, displayTitle))
	if body != "" {
		htmlOutput.WriteString(fmt.Sprintf("<pre class=\"log-content\">%s</pre>\n", html.EscapeString(body)))
	}
	htmlOutput.WriteString("</details>\n")
}
func SaveHTMLReport() {
	htmlOnce.Do(func() {})
	var toc strings.Builder
	toc.WriteString("<h2>Table of Contents</h2><ul>")
	for _, item := range sectionIDs {
		parts := strings.SplitN(item, "|", 2)
		if len(parts) == 2 {
			id := parts[0]
			title := parts[1]
			toc.WriteString(fmt.Sprintf("<li><a href=\"#%s\">%s</a></li>", id, html.EscapeString(title)))
		}
	}
	toc.WriteString("</ul>")
	searchBox := `
	<h2>Search</h2>
	<input type="text" id="searchBox" placeholder="Search logs, KCS, AI analysis..." style="width:100%;padding:8px;font-size:14px;margin-bottom:10px;">
	<script>
	function highlight(term) {
	  let sections = document.querySelectorAll("details");
	  sections.forEach(sec => {
	    let pre = sec.querySelector("pre");
	    if (!pre) return;
	    let text = pre.innerText;
	    if (term === "") {
	      pre.innerHTML = text;
	      sec.open = false;
	    } else {
	      let regex = new RegExp("(" + term + ")", "gi");
	      let hasMatch = regex.test(text);
	      pre.innerHTML = text.replace(regex, '<mark>$1</mark>');
	      sec.open = hasMatch;
	    }
	  });
	}
	document.getElementById("searchBox").addEventListener("input", function() {
	  highlight(this.value);
	});
	</script>
	`
	// 整个 HTML 页面
	content := "<html><head><meta charset='UTF-8'><title>Inceptor-Lite Report</title>" +
		"<style>body{font-family:system-ui,Segoe UI,Helvetica,Arial,sans-serif;line-height:1.35;background:#f9f9f9}h1,h2{margin:16px 0 8px}pre{background:#f6f8fa;padding:10px;border-radius:6px;overflow:auto;font-size:13px;white-space:pre-wrap}summary{cursor:pointer;font-size:16px;margin:4px 0}details{margin:8px 0;border:1px solid #ddd;border-radius:6px;padding:4px;background:#fff}a{text-decoration:none;color:#0366d6}mark{background:yellow;font-weight:bold}</style>" +
		"</head><body><h1>Inceptor-Lite Troubleshooting Report</h1>" +
		toc.String() + searchBox + htmlOutput.String() + "</body></html>"
	out := filepath.Join(reportDir, "report.html")
	_ = os.WriteFile(out, []byte(content), 0644)
	fmt.Printf("HTML report saved: %s\n", out)
}
