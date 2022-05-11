package browserControl

import (
	"github.com/chromedp/chromedp"
	"github.com/goph/emperror"
	"github.com/je4/bremote/v2/browser"
	"github.com/op/go-logging"
	"io/ioutil"
	"net/url"
	"path/filepath"
	"sync"
	"time"
)

type BrowserControl struct {
	browser      *browser.Browser
	homeUrl      *url.URL
	opts         map[string]any
	timeout      time.Duration
	logger       *logging.Logger
	lastLog      time.Time
	lastLogMutex sync.RWMutex
	stop         chan any
}

func NewBrowserControl(homeUrl *url.URL, opts map[string]any, timeout time.Duration, logger *logging.Logger) (*BrowserControl, error) {
	bc := &BrowserControl{
		homeUrl: homeUrl,
		opts:    opts,
		timeout: timeout,
		logger:  logger,
	}
	return bc, nil
}

func (bc *BrowserControl) log(string, ...any) {
	bc.lastLogMutex.Lock()
	bc.lastLog = time.Now()
	defer bc.lastLogMutex.Unlock()
}

func (bc *BrowserControl) Start() error {
	var err error
	bc.browser, err = browser.NewBrowser(bc.opts, bc.logger, bc.log)
	if err != nil {
		return emperror.Wrap(err, "cannot create browser instance")
	}
	// ensure that the browser process is started
	if err := bc.browser.Run(); err != nil {
		return emperror.Wrap(err, "cannot run browser")
	}

	path := filepath.Join(bc.browser.TempDir, "DevToolsActivePort")
	bs, err := ioutil.ReadFile(path)
	if err != nil {
		bc.logger.Panicf("error reading DevToolsActivePort: %v", err)
	}
	//	lines := bytes.Split(bs, []byte("\n"))
	bc.logger.Debugf("DevToolsActivePort:\n%v", string(bs))
	tasks := chromedp.Tasks{
		chromedp.Navigate(bc.homeUrl.String()),
		//		browser.MouseClickXYAction(2,2),
	}
	err = bc.browser.Tasks(tasks)
	if err != nil {
		bc.logger.Errorf("could not navigate: %v", err)
	}

	bc.stop = make(chan any)

	go bc.mainLoop()

	return nil
}

func (bc *BrowserControl) Shutdown() {
	close(bc.stop)
	if bc.browser == nil {
		return
	}
	bc.browser.Close()
}

func (bc *BrowserControl) mainLoop() {
	for {
		select {
		case <-time.After(time.Second * 3):
		case <-bc.stop:
			return
		}
		if !bc.browser.IsRunning() {
			bc.browser.Startup()
			tasks := chromedp.Tasks{
				chromedp.Navigate(bc.homeUrl.String()),
				//		browser.MouseClickXYAction(2,2),
			}
			if err := bc.browser.Tasks(tasks); err != nil {
				bc.logger.Errorf("could not navigate: %v", err)
			}
		}
		bc.lastLogMutex.RLock()
		timeout := time.Now().After(bc.lastLog.Add(bc.timeout))
		bc.lastLogMutex.RUnlock()
		if timeout {
			tasks := chromedp.Tasks{
				chromedp.Navigate(bc.homeUrl.String()),
				//		browser.MouseClickXYAction(2,2),
			}
			if err := bc.browser.Tasks(tasks); err != nil {
				bc.logger.Errorf("could not navigate: %v", err)
			}
		}
	}
}
