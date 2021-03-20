package units

import (
	"context"
	"log"
	"time"

	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

// TakeScreenshot ...
func TakeScreenshot(url string, id int) (*[]byte, error) {
	start := time.Now()

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

	defer acancel()
	defer cancel()

	var buf []byte

	if err := chromedp.Run(ctx, fullScreenshot(url, 1280, 960, 50, &buf)); err != nil {
		log.Println("chromedp.Run: ", err)
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
