package utils

import (
    "bytes"
    "fmt"
    "os/exec"
    "strings"
)

func RunCommand(name string, args ...string) (string, error) {
    cmd := exec.Command(name, args...)
    var out bytes.Buffer
    cmd.Stdout = &out
    cmd.Stderr = &out
    err := cmd.Run()
    if err != nil {
        return "", fmt.Errorf("%s", out.String())
    }
    return out.String(), nil
}

func RunShellCommand(command string) (string, error) {
    cmd := exec.Command("bash", "-c", command)
    var out bytes.Buffer
    cmd.Stdout = &out
    cmd.Stderr = &out
    err := cmd.Run()
    if err != nil {
        return "", fmt.Errorf("command failed: %s", out.String())
    }
    return out.String(), nil
}

func TailLines(s string, n int) string {
    lines := strings.Split(s, "\n")
    if len(lines) <= n {
        return s
    }
    return strings.Join(lines[len(lines)-n:], "\n")
}

func HighlightKeywords(log string, keywords []string) string {
    for _, kw := range keywords {
        log = strings.ReplaceAll(log, kw, "\033[31m"+kw+"\033[0m")
    }
    return log
}