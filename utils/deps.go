package utils

import (
    "fmt"
    "os/exec"
)

func CheckDependencies(commands []string) error {
    for _, cmd := range commands {
        if _, err := exec.LookPath(cmd); err != nil {
            return fmt.Errorf("missing dependency: %s", cmd)
        }
    }
    return nil
}