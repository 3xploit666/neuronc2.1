package utils

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

func SaveScreenshot(agentID, b64 string) error {
	data, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		return err
	}

	screenshotDir := "screenshots"
	if err := os.MkdirAll(screenshotDir, 0755); err != nil {
		return err
	}

	filename := fmt.Sprintf("%s/%s-%d.png", screenshotDir, agentID, time.Now().Unix())
	if err := os.WriteFile(filepath.Clean(filename), data, 0644); err != nil {
		return err
	}

	return nil
}
