package main

import (
	"fmt"
	"inceptor-lite/ai"
	"inceptor-lite/kcs"
	"inceptor-lite/omc"
	"inceptor-lite/ui"
	"inceptor-lite/utils"
	"os"
	"strings"
	"time"
)

func getOperatorConfigs() map[string]omc.OperatorConfig {
	return map[string]omc.OperatorConfig{
		"authentication": {
			Name: "authentication",
			LogSources: []omc.LogSource{
				{Namespace: "openshift-authentication", Lines: 100},
			},
			SearchPatterns: []string{
				"OAuthServerConfigObservationDegraded",
				"OAuthServerRouteEndpointAccessibleControllerDegraded",
				"ProxyConfigControllerDegraded",
				"APIServerDeploymentDegraded",
				"RouteStatusDegraded",
			},
		},
		"image-registry": {
			Name: "image-registry",
			LogSources: []omc.LogSource{
				{Namespace: "openshift-image-registry", Lines: 100},
			},
			SearchPatterns: []string{"ImageRegistryDegraded", "DeploymentFailed"},
		},
		"dns": {
			Name: "dns",
			LogSources: []omc.LogSource{
				{Namespace: "openshift-dns", Lines: 100},
			},
			SearchPatterns: []string{"DNSDegraded", "DNSUnavailable"},
		},
		"ingress": {
			Name: "ingress",
			LogSources: []omc.LogSource{
				{Namespace: "openshift-ingress", Lines: 100},
			},
			SearchPatterns: []string{"IngressControllerDegraded", "IngressUnavailable"},
		},
		"console": {
			Name: "console",
			LogSources: []omc.LogSource{
				{Namespace: "openshift-console", Lines: 100},
			},
			SearchPatterns: []string{"ConsoleDegraded", "ConsoleUnavailable"},
		},
	}
}
func getOperatorsToCheck(operatorConfigs map[string]omc.OperatorConfig) []string {
	if len(os.Args) > 1 {
		return []string{os.Args[1]}
	}
	var ops []string
	for op := range operatorConfigs {
		ops = append(ops, op)
	}
	return ops
}
func main() {
	fmt.Println("\u231B Checking dependencies...") // :hourglass:
	if err := utils.CheckDependencies([]string{"omc", "zenity", "jq"}); err != nil {
		fmt.Printf("\u274C Dependency check failed: %v\n", err) // :x:
		os.Exit(1)
	}
	username, password := ui.PromptCredentials()
	fmt.Printf("\u2705 Welcome, %s\n", username) // :white_check_mark:
	if err := ui.DeleteAll(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	path, err := ui.SelectMustGather()
	if err != nil {
		fmt.Printf("\u274C Failed to select must-gather folder: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("\U0001F4C1 Selected path: %s\n", path) // :file_folder:
	omc.SetReportDir(path)
	if err := omc.UseMustGather(path); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	var combinedContext strings.Builder
	// Node 检查
	nodesText, nodeErr := omc.GetNodeStatus()
	if nodeErr == nil && strings.TrimSpace(nodesText) != "" {
		combinedContext.WriteString("### Node Status\n" + nodesText + "\n")
		nodeTerms := omc.InspectProblematicNodesAndCollectSearchTerms()
		if len(nodeTerms) > 0 {
			combinedContext.WriteString("### Problematic Node Conditions\n" + strings.Join(nodeTerms, "\n") + "\n")
		}
	}
	operatorConfigs := getOperatorConfigs()
	for _, operator := range getOperatorsToCheck(operatorConfigs) {
		// 处理前提示
		fmt.Printf("\r\u23F3 Checking Operator: %s ...", operator) // :hourglass_flowing_sand:
		config := operatorConfigs[operator]
		// Operator degraded 条件
		degraded := omc.ExtractOperatorConditions(config.Name)
		if len(degraded) > 0 {
			text := strings.Join(degraded, "\n")
			fmt.Printf("\n[Operator Keyword] %s\n", text)
			omc.AppendSection(fmt.Sprintf("Operator %s - Degraded", config.Name), text)
			kcs.SearchAndAppend(username, password, text, fmt.Sprintf("Operator %s - Degraded", config.Name))
			_, _ = ai.AnalyzeWithSource(text, fmt.Sprintf("Operator %s - Degraded", config.Name))
			combinedContext.WriteString(fmt.Sprintf("### Operator %s Degraded\n%s\n", operator, text))
		}
		// Operator 日志
		analyzer := omc.OperatorAnalyzer{Config: config}
		_ = analyzer.CollectLogs()
		logSearch := analyzer.BuildLogSearchString()
		if logSearch != "" {
			fmt.Printf("\n[Logs Keyword] %s\n", logSearch)
			omc.AppendSection(fmt.Sprintf("Operator %s - Logs", config.Name), logSearch)
			kcs.SearchAndAppend(username, password, logSearch, fmt.Sprintf("Operator %s - Logs", config.Name))
			_, _ = ai.AnalyzeWithSource(logSearch, fmt.Sprintf("Operator %s - Logs", config.Name))
			combinedContext.WriteString(fmt.Sprintf("### Operator %s Logs\n%s\n", operator, logSearch))
		}
		// 完成提示
		fmt.Printf("\r\u2705 Operator %s done\n", operator) // :white_check_mark:
		time.Sleep(1 * time.Second)
	}
	// Namespace-only 回退
	if nodeErr != nil {
		fallbackAnalyzer := omc.OperatorAnalyzer{}
		omc.CollectNamespaceOnlyLogsAndBuildSearch(&fallbackAnalyzer)
		logSearch := fallbackAnalyzer.BuildLogSearchString()
		if logSearch != "" {
			combinedContext.WriteString("### Namespace-Only Log Search\n" + logSearch + "\n")
		}
	}
	// 全局综合分析
	fullContext := combinedContext.String()
	if fullContext != "" {
		omc.AppendSection("Global Context", fullContext)
		kcs.SearchAndAppend(username, password, fullContext, "Global Context")
		_, _ = ai.AnalyzeWithSource(fullContext, "Global Analysis")
	}
	// 生成报告
	omc.SaveHTMLReport()
}
