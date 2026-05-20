package main

import "os/exec"

func capturePane(paneID string) (string, error) {
	out, err := exec.Command("tmux", "capture-pane", "-t", paneID, "-p").Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}
