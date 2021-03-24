package units

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTakeScreenshot(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "so testy :)")
	}))
	defer s.Close()

	chromePath := "/usr/bin/google-chrome-stable"
	userDataDir := t.TempDir()
	screenshotPath := t.TempDir()

	screenshotter := NewScreenshotter(chromePath, userDataDir, screenshotPath, 5*time.Second)
	_, err := screenshotter.TakeScreenshot(s.URL+"screnshotter", 1)
	assert.Nil(t, err)

	imagePathWant := screenshotPath + "/image_" + fmt.Sprint(1) + ".png"
	imagePath, err := screenshotter.TakeScreenshot(s.URL+"screnshotter", 1)
	assert.Nil(t, err)
	assert.Equal(t, imagePathWant, imagePath)

	// Failure scenario - no path to Chrome
	screenshotter = NewScreenshotter("", userDataDir, screenshotPath, 5*time.Second)
	_, err = screenshotter.TakeScreenshot(s.URL+"screnshotter", 1)
	assert.NotNil(t, err)

	// Failure scenario - timeout
	screenshotter = NewScreenshotter(chromePath, userDataDir, screenshotPath, 1*time.Millisecond)
	_, err = screenshotter.TakeScreenshot(s.URL+"screnshotter", 1)
	assert.NotNil(t, err)
}
