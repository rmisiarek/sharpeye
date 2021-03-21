package units

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"time"
)

// TODO: read from config or just try to find it
var chromePath = "/usr/bin/google-chrome-stable"

// TakeScreenshot takes a screenshot of the website given as a parameter.
// It returns string with path to saved screenshot if succeed or empty
// string when it fails.
func TakeScreenshot(url string, id int) (string, error) {
	screenshotFilePath := screenshotFilePath(id)

	var chromeOpts = []string{
		"--headless",
		"--no-sandbox", // docker
		"--disable-gpu",
		"--incognito",
		"--no-first-run",
		"--mute-audio",
		"--hide-scrollbars",
		"--disable-notifications",
		"--disable-crash-reporter",
		"--ignore-certificate-errors",
		"--no-default-browser-check",
		"--user-data-dir=screens3", // TODO: change it
		"--disable-infobars",
		"--disable-sync",
		// "--user-agent=",
		"--window-size=1280,960",
		"--screenshot=" + screenshotFilePath,
	}

	chromeOpts = append(chromeOpts, url)

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(30*time.Second))
	defer cancel()

	cmd := exec.CommandContext(ctx, chromePath, chromeOpts...)
	if err := cmd.Start(); err != nil {
		killCmd(cmd)
		return "", errors.New("in cmd.Start()")
	}

	if err := cmd.Wait(); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			killCmd(cmd)
			return "", errors.New("timed out")
		}
		killCmd(cmd)
		return "", errors.New("in cmd.Wait()")
	}

	killCmd(cmd)
	return screenshotFilePath, nil
}

func killCmd(cmd *exec.Cmd) {
	cmd.Process.Release()
	cmd.Process.Kill()
}

func screenshotFilePath(id int) string {
	return "screens3/image_" + fmt.Sprint(id) + ".png"
}
