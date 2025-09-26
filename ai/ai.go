package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"inceptor-lite/omc"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

type AIResponse struct {
	Choices []struct {
		Text string `json:"text"`
	} `json:"choices"`
}

func AnalyzeWithSource(context, source string) (string, error) {
	ctx := strings.TrimSpace(context)
	if ctx == "" {
		return "", nil
	}
	prompt := fmt.Sprintf(
		`Context Source: %s
%s
Please analyze and provide:
1) Likely root causes
2) Step-by-step troubleshooting plan (with specific omc/oc commands and must-gather files)
3) Suggested remediations
4) Reference links: Red Hat docs, KCS, upstream resources

Output in clear Markdown format with sections and bullet points.`, source, ctx)
	payload := map[string]interface{}{
		"model":       "ibm-granite/granite-3.3-8b-instruct",
		"prompt":      prompt,
		"max_tokens":  3000,
		"temperature": 0,
	}

	data, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST",
		"https://granite-3-3-8b-instruct--apicast-production.apps.int.stc.ai.prod.us-east-1.aws.paas.redhat.com:443/v1/completions",
		bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	token := os.Getenv("AI_AUTH_TOKEN")
	//	if token == "" {

	//	}
	req.Header.Set("Authorization", "Bearer "+token)
	client := &http.Client{Timeout: 90 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("AI request failed: %w", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("AI returned %d: %s", resp.StatusCode, string(body))
	}
	var res AIResponse
	if json.Unmarshal(body, &res) != nil || len(res.Choices) == 0 {
		return "", fmt.Errorf("AI invalid response")
	}
	answer := strings.TrimSpace(res.Choices[0].Text)
	omc.AppendSection("AI Analysis - "+source, answer)
	return answer, nil
}
