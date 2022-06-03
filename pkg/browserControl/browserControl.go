package browserControl

import (
	"context"
	"github.com/chromedp/cdproto/fetch"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"github.com/goph/emperror"
	"github.com/je4/bremote/v2/browser"
	"github.com/op/go-logging"
	"io/ioutil"
	"net/url"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type BrowserControl struct {
	browser       *browser.Browser
	homeUrl       *url.URL
	opts          map[string]any
	timeout       int64
	logger        *logging.Logger
	lastLog       int64
	lastLogMutex  sync.RWMutex
	stop          chan any
	allowedPrefix string
	taskDelay     time.Duration
}

func NewBrowserControl(allowedPrefix string, homeUrl *url.URL, opts map[string]any, timeout, taskDelay time.Duration, logger *logging.Logger) (*BrowserControl, error) {
	bc := &BrowserControl{
		allowedPrefix: allowedPrefix,
		homeUrl:       homeUrl,
		opts:          opts,
		timeout:       int64(timeout.Seconds()),
		taskDelay:     taskDelay,
		logger:        logger,
	}
	return bc, nil
}

func (bc *BrowserControl) log(str string, evs ...any) {
	atomic.StoreInt64(&bc.lastLog, time.Now().Unix())
	if len(evs) == 0 {
		return
	}
	ev := evs[0]
	switch ev := ev.(type) {
	case *network.EventRequestWillBeSent:
		// must not block
		go func(ctx context.Context, ev *network.EventRequestWillBeSent) {
			if !strings.HasPrefix(ev.DocumentURL, bc.allowedPrefix) {
				bc.logger.Infof("forbidden URL: %s", ev.DocumentURL)
				tasks := chromedp.Tasks{
					chromedp.Navigate(bc.homeUrl.String()),
					//		browser.MouseClickXYAction(2,2),
				}
				if err := bc.browser.Tasks(tasks); err != nil {
					bc.logger.Errorf("could not navigate: %v", err)
				}
			}
		}(bc.browser.TaskCtx, ev)
	}
	//bc.logger.Debugf("%s - %v", str, param)
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
		fetch.Disable(),
	}
	time.Sleep(bc.taskDelay)
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
		llog := atomic.LoadInt64(&bc.lastLog)
		timeout := time.Now().Unix()-llog > bc.timeout
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
