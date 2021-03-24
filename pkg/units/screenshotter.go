package units

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"time"
)

// Screenshotter holds base data to take screenshot by Chrome headless API.
// ChromePath - path to Chrome executable.
// UserDataPath - path to directory where Chrome data will be stored.
// ScreenshotPath - path to directory where screenshots will be stored.
// Timeout - duration for timeout.
type Screenshotter struct {
	ChromePath     string
	UserDataPath   string
	ScreenshotPath string
	Timeout        time.Duration
}

// NewScreenshotter it's a constructor for Screenshotter base struct.
// See Screenshotter struct for parameters description.
func NewScreenshotter(
	chromePath, userDataPath, screenshotPath string, timeout time.Duration,
) *Screenshotter {

	return &Screenshotter{
		ChromePath:     chromePath,
		UserDataPath:   userDataPath,
		ScreenshotPath: screenshotPath,
		Timeout:        timeout,
	}
}

// TakeScreenshot takes a screenshot of the website given as a parameter.
// It returns string with path to saved screenshot if succeed or empty
// string when it fails.
func (s *Screenshotter) TakeScreenshot(url string, id int) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(s.Timeout))
	defer cancel()

	opts := []string{
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
		"--disable-infobars",
		"--disable-sync",
		"--window-size=1280,960",
		"--user-data-dir=" + s.UserDataPath,
		// "--user-agent=", // TODO: verify it needed or useful
	}

	imagePath := s.ScreenshotPath + "/image_" + fmt.Sprint(id) + ".png"
	opts = append(opts, "--screenshot="+imagePath)
	opts = append(opts, url)

	cmd := exec.CommandContext(ctx, s.ChromePath, opts...)
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
	return imagePath, nil
}

func killCmd(cmd *exec.Cmd) {
	if cmd.Path != "" {
		cmd.Process.Release()
		cmd.Process.Kill()
	}
}
