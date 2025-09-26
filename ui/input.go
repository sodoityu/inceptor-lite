package ui

import (
    "bufio"
    "fmt"
    "os"
    "strings"

    "golang.org/x/term"
)

func PromptCredentials() (string, string) {
    reader := bufio.NewReader(os.Stdin)

    fmt.Print("Enter your username: ")
    username, _ := reader.ReadString('\n')
    username = strings.TrimSpace(username)

    fmt.Print("Enter your password: ")
    bytePassword, _ := term.ReadPassword(int(os.Stdin.Fd()))
    fmt.Println()

    return username, string(bytePassword)
}