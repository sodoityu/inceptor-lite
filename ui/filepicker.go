package ui

import (
	"fmt"
	"inceptor-lite/utils"
	"os"
	"os/exec"
	"strings"
)

func DeleteAll() error {
	cmd := exec.Command("omc", "mg", "delete", "-a")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	fmt.Println("Executing: omc mg delete -a ...")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to run omc mg delete -a: %w", err)
	}
	fmt.Println("omc mg delete -a executed successfully")
	return nil
}

func SelectMustGather() (string, error) {
	utils.RunCommand("omc", "delete", "-a")
	out, err := exec.Command("zenity", "--file-selection", "--title=Select the must-gather/inspect file", "--directory").Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}
