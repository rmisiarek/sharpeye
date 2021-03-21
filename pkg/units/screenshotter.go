package units

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os/exec"
	"time"

	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

// TakeScreenshot ...
func TakeScreenshot(url string, id int) (*[]byte, error) {
	start := time.Now()
	fmt.Println("Start TakeScreenshot - ", start)
	options := []chromedp.ExecAllocatorOption{}
	options = append(options, chromedp.DefaultExecAllocatorOptions[:]...)
	// options = append(options, chromedp.UserAgent(chrome.UserAgent))
	// options = append(options, chromedp.DisableGPU)
	// options = append(options, chromedp.Headless)
	// options = append(options, chromedp.NoSandbox)
	// options = append(options, chromedp.NoFirstRun)
	// options = append(options, chromedp.NoDefaultBrowserCheck)
	options = append(options, chromedp.WindowSize(1280, 960))
	options = append(options, chromedp.Flag("ignore-certificate-errors", true))

	actx, acancel := chromedp.NewExecAllocator(context.Background(), options...)
	ctx, cancel := chromedp.NewContext(actx)
	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)

	defer acancel()
	defer cancel()

	var buf []byte

	if err := chromedp.Run(ctx, fullScreenshot(url, 1280, 960, 50, &buf)); err != nil {
		log.Println("chromedp.Run: ", err)
	}

	path := "/home/raf/go/src/sharpeye/screens3/image_" + fmt.Sprint(id) + ".png"
	err := ioutil.WriteFile(path, buf, 0644)
	if err != nil {
		log.Fatal(err)
	}

	elapsed := time.Since(start)
	log.Printf("TakeScreenshot took %s", elapsed)

	return &buf, nil
}

func fullScreenshot(
	urlstr string, width int64, height int64, quality int64, res *[]byte,
) chromedp.Tasks {

	return chromedp.Tasks{
		chromedp.Navigate(urlstr),
		chromedp.ActionFunc(func(ctx context.Context) error {
			err := emulation.SetDeviceMetricsOverride(width, height, 1, false).
				WithScreenOrientation(&emulation.ScreenOrientation{
					Type:  emulation.OrientationTypePortraitPrimary,
					Angle: 0,
				}).Do(ctx)

			if err != nil {
				return err
			}

			*res, err = page.CaptureScreenshot().
				WithQuality(quality).
				WithClip(&page.Viewport{
					Width:  float64(width),
					Height: float64(height),
					Scale:  1,
				}).Do(ctx)

			if err != nil {
				return err
			}

			return nil
		}),
	}
}

func TakeScr(url string, id int) (string, error) {
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

	chromePath := "/usr/bin/google-chrome-stable"
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
