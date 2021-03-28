package core

import (
	"log"
	"net/http"

	"github.com/rmisiarek/sharpeye/pkg/db"
	"github.com/rmisiarek/sharpeye/pkg/units"
)

// Processor ...
type Processor struct {
	ID     int
	URL    string
	DB     *db.Session
	Client *http.Client
}

var (
	client        = &http.Client{}
	session       = &db.Session{}
	screenshotter = &units.Screenshotter{}
)

func init() {
	client = units.NewHTTPClient()
	screenshotter = units.NewScreenshotter(
		"/usr/bin/google-chrome-stable",
		"/home/raf/go/src/sharpeye/debug/user-data",
		"/home/raf/go/src/sharpeye/debug/screens",
		15,
	)
	session = db.InitDB()
}

// NewProcessor ...
func NewProcessor(id int, url string) *Processor {
	return &Processor{
		ID:     id,
		URL:    url,
		DB:     session,
		Client: client,
	}
}

// Sharpeye ...
func (p *Processor) Sharpeye() {
	proberResult, tlsResult := p.probeHost()
	imagePath := p.takeScreenshot()

	d := db.SharpeyeDocument{
		ID:        p.ID,
		URL:       p.URL,
		ImagePath: imagePath,
		Prober:    *proberResult,
		TLS:       *tlsResult,
	}

	p.DB.Save(&d)
}

func (p *Processor) probeHost() (*units.ProberResponse, *units.ProberTLSResponse) {
	prober, tls, err := units.ProbeHost(p.Client, p.URL)
	if err != nil {
		printError("prober", p.URL, err.Error())
	} else {
		printOk("prober", p.URL)
	}

	return prober, tls
}

func (p *Processor) takeScreenshot() string {
	path, err := screenshotter.TakeScreenshot(p.URL, p.ID)
	if err != nil {
		printError("screenshotter", p.URL, err.Error())
	} else {
		printOk("screenshotter", p.URL)
	}

	return path
}

func printOk(unit, url string) {
	log.Printf("| %s: %s | %s \n", Green(unit), Green("OK"), url)
}

func printError(unit, url, err string) {
	log.Printf("| %s: %s | %s \n", Red(unit), Red(err), url)
}
