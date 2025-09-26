package omc

import (
	"encoding/json"
	"fmt"
	"html"
	"inceptor-lite/utils"
	"regexp"
	"strings"
)

type LogSource struct {
	Namespace string
	PodLabel  string
	Container string
	Lines     int
}
type OperatorConfig struct {
	Name           string
	LogSources     []LogSource
	SearchPatterns []string
}
type OperatorAnalyzer struct {
	Config OperatorConfig
	Logs   []string
}

var (
	logKeywords = []string{"error", "degraded", "timeout", "fail", "failed", "crash", "unavailable", "denied"}
	logContext  = 2
)

func UseMustGather(path string) error {
	out, err := utils.RunCommand("omc", "use", path)
	AppendSection("Must-Gather Loaded", out)
	return err
}
func GetNodeStatus() (string, error) {
	out, err := utils.RunCommand("omc", "get", "nodes")
	if err == nil && strings.TrimSpace(out) != "" {
		AppendSection("Node Status", out)
	}
	return out, err
}
func InspectProblematicNodesAndCollectSearchTerms() []string {
	raw, _ := utils.RunCommand("omc", "get", "nodes", "-o", "json")
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	type Condition struct {
		Type    string `json:"type"`
		Status  string `json:"status"`
		Reason  string `json:"reason"`
		Message string `json:"message"`
	}
	type Node struct {
		Metadata struct{ Name string } `json:"metadata"`
		Status   struct {
			Conditions []Condition `json:"conditions"`
		} `json:"status"`
	}
	type NodeList struct{ Items []Node }
	var list NodeList
	if json.Unmarshal([]byte(raw), &list) != nil {
		return nil
	}
	var problems []string
	for _, n := range list.Items {
		for _, c := range n.Status.Conditions {
			isProblem := (c.Type == "Ready" && c.Status != "True") || (c.Type != "Ready" && c.Status == "True")
			if isProblem && strings.ToLower(c.Status) != "false" {
				line := fmt.Sprintf("[%s] %s=%s %s %s", n.Metadata.Name, c.Type, c.Status, c.Reason, c.Message)
				problems = append(problems, line)
			}
		}
	}
	if len(problems) > 0 {
		AppendSection("Problematic Nodes", strings.Join(problems, "\n"))
	}
	return problems
}
func ExtractOperatorConditions(operator string) []string {
	out, _ := utils.RunCommand("omc", "get", "co", operator, "-o", "json")
	if strings.TrimSpace(out) == "" {
		return nil
	}
	type Condition struct {
		Type    string `json:"type"`
		Status  string `json:"status"`
		Message string `json:"message"`
	}
	type Status struct {
		Conditions []Condition `json:"conditions"`
	}
	type CO struct{ Status Status }
	var co CO
	if json.Unmarshal([]byte(out), &co) != nil {
		return nil
	}
	var degraded []string
	for _, c := range co.Status.Conditions {
		if (c.Type == "Degraded" || c.Type == "Progressing") && c.Status == "True" {
			degraded = append(degraded, c.Message)
		}
	}
	return degraded
}
func (oa *OperatorAnalyzer) CollectLogs() error {
	for _, src := range oa.Config.LogSources {
		pods, _ := utils.RunCommand("omc", "get", "pods", "-n", src.Namespace, "-o", "name")
		for _, p := range strings.Split(strings.TrimSpace(pods), "\n") {
			if p == "" {
				continue
			}
			AppendSection(fmt.Sprintf("Pod: %s (ns:%s)", p, src.Namespace), "")
			containers, _ := utils.RunCommand("omc", "get", p, "-n", src.Namespace, "-o", "jsonpath={.spec.containers[*].name}")
			for _, c := range strings.Fields(containers) {
				log, _ := utils.RunCommand("omc", "logs", p, "-n", src.Namespace, "-c", c)
				AppendSection(fmt.Sprintf("Container %s Logs", c), log)
				matched := ExtractKeywordsFromLogs(log, logKeywords)
				for _, line := range matched {
					line = stripANSI(line)
					if len(line) > 500 {
						line = line[:500]
					}
					oa.Logs = append(oa.Logs, fmt.Sprintf("[%s/%s] %s", p, c, line))
					writeHTMLLine(html.EscapeString(fmt.Sprintf("[HIT] %s/%s: %s", p, c, line)))
				}
			}
		}
	}
	oa.Logs = unique(oa.Logs)
	return nil
}
func ExtractKeywordsFromLogs(logs string, keywords []string) []string {
	var found []string
	lines := strings.Split(logs, "\n")
	for i, line := range lines {
		l := strings.ToLower(line)
		match := false
		for _, kw := range keywords {
			if strings.Contains(l, strings.ToLower(kw)) {
				match = true
				break
			}
		}
		if match {
			start := i - logContext
			if start < 0 {
				start = 0
			}
			end := i + logContext
			if end >= len(lines) {
				end = len(lines) - 1
			}
			found = append(found, lines[start:end+1]...)
		}
	}
	return unique(found)
}
func (oa *OperatorAnalyzer) BuildLogSearchString() string {
	if len(oa.Logs) == 0 {
		return ""
	}
	full := stripANSI(strings.Join(oa.Logs, " "))
	if len(full) > 500 {
		full = full[:500]
	}
	return full
}
func CollectNamespaceOnlyLogsAndBuildSearch(an *OperatorAnalyzer) {
	namespaces, _ := utils.RunCommand("omc", "get", "namespace", "-o", "name")
	for _, ns := range strings.Split(strings.TrimSpace(namespaces), "\n") {
		if ns == "" {
			continue
		}
		pods, _ := utils.RunCommand("omc", "get", "pods", "-n", ns, "-o", "name")
		for _, p := range strings.Split(strings.TrimSpace(pods), "\n") {
			if p == "" {
				continue
			}
			containers, _ := utils.RunCommand("omc", "get", p, "-n", ns, "-o", "jsonpath={.spec.containers[*].name}")
			for _, c := range strings.Fields(containers) {
				log, _ := utils.RunCommand("omc", "logs", p, "-n", ns, "-c", c)
				matched := ExtractKeywordsFromLogs(log, logKeywords)
				for _, line := range matched {
					line = stripANSI(line)
					if len(line) > 500 {
						line = line[:500]
					}
					an.Logs = append(an.Logs, fmt.Sprintf("[%s/%s/%s] %s", ns, p, c, line))
				}
			}
		}
	}
	an.Logs = unique(an.Logs)
}
func unique(input []string) []string {
	seen := make(map[string]struct{}, len(input))
	out := make([]string, 0, len(input))
	for _, s := range input {
		if _, ok := seen[s]; !ok {
			seen[s] = struct{}{}
			out = append(out, s)
		}
	}
	return out
}
func stripANSI(s string) string {
	re := regexp.MustCompile(`\x1b\[[0-9;]*m`)
	return re.ReplaceAllString(s, "")
}
