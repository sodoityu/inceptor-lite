package kcs

import (
	"encoding/base64"
	"fmt"
	"inceptor-lite/omc"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

func buildKCSURL(searchString string) string {
	base := "https://api.access.redhat.com/support/search/kcs"
	query := url.QueryEscape("*" + searchString + "*")
	return fmt.Sprintf("%s?fq=documentKind:(\"Solution\")&q=%s&rows=3&start=0", base, query)
}
func extractSolutionLinks(apiResp string) []string {
	re := regexp.MustCompile(`https://access.redhat.com/solutions/\d+`)
	return re.FindAllString(apiResp, -1)
}

// SearchSolutions 调用 KCS API，返回匹配到的链接
func SearchSolutions(username, password, searchString string) ([]string, error) {
	if strings.TrimSpace(searchString) == "" {
		return nil, fmt.Errorf("no search keyword provided")
	}
	apiURL := buildKCSURL(searchString)
	auth := base64.StdEncoding.EncodeToString([]byte(username + ":" + password))
	req, _ := http.NewRequest("GET", apiURL, nil)
	req.Header.Add("Authorization", "Basic "+auth)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("KCS API returned status: %s", resp.Status)
	}
	bodyBytes, _ := io.ReadAll(resp.Body)
	body := string(bodyBytes)
	links := extractSolutionLinks(body)
	return links, nil
}

// SearchAndAppend 调用 SearchSolutions 并把结果写到 HTML 报告里
func SearchAndAppend(username, password, keyword, source string) {
	links, err := SearchSolutions(username, password, keyword)
	title := fmt.Sprintf("KCS Solutions - %s [Keyword: %s]", source, keyword)
	if err != nil {
		omc.AppendSection(title, fmt.Sprintf("KCS search failed: %v", err))
		return
	}
	if len(links) == 0 {
		omc.AppendSection(title, "No KCS solutions found")
		return
	}
	var builder strings.Builder
	builder.WriteString("Found potential KCS solutions:\n")
	for _, link := range links {
		builder.WriteString("- " + link + "\n")
	}
	// 在 HTML 报告里追加完整结果
	omc.AppendSection(title, builder.String())
}
