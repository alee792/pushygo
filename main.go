package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

var c *chromedp.CDP
var cancelChan chan struct{}
var reloadActive = false

func init() {
	log.SetLevel(log.DebugLevel)
}

// Chrome is a struct to hold context and CDP
type Chrome struct {
	Context *context.Context
	CDP     *chromedp.CDP
}

// YTRequest holds info relevant to a YouTube request
type YTRequest struct {
	ID string `json:"id,omitempty"`
}

// URLRequest holds info relevant to a URL request
type URLRequest struct {
	URL      string `json:"url"`
	Protocol string `json:"protocol"`
	Opts     `json:"opts"`
}

// Opts provides generic configs for a CDP task
type Opts struct {
	Fullscreen    int `json:"fullscreen"`
	HideScrollbar int `json:"hideScrollbar"`
	ReloadSeconds int `json:"reloadSeconds"`
}

// GetYT handles GET requests to the /youtube/{id} endpoint
func (c *Chrome) GetYT(w http.ResponseWriter, r *http.Request) {

	fmt.Fprintf(w, "Getting Youtube.")
	params := mux.Vars(r)

	url := fmt.Sprintf("https://youtube.com/embed/%s?autoplay=1", params["id"])
	err := c.navigate(url)
	if err != nil {
		log.Error(err)
	}
}

// GetKibana handles GET requests to the /kibana/ endpoint
func (c *Chrome) GetKibana(w http.ResponseWriter, r *http.Request) {

	fmt.Fprintf(w, "Getting Kibana.")
	url := ""

	err := c.navigate(url)
	if err != nil {
		log.Error(err)
	}
}

// GetURL handles GET requests to the /url/{http/s} endpoint
func (c *Chrome) GetURL(w http.ResponseWriter, r *http.Request) {

	params := mux.Vars(r)
	fmt.Fprintf(w, "Getting URL %s.", params["url"])
	protocol := params["protocol"]
	if protocol == "" {
		protocol = "https"
	}
	url := fmt.Sprintf("%s://%v", protocol, params["url"])
	err := c.navigate(url)
	if err != nil {
		log.Error(err)
	}
}

// PostURL handles POST requests to the /url/ endpoint
func (c *Chrome) PostURL(w http.ResponseWriter, r *http.Request) {

	log.Info("POST navigating to url")

	var urlRequest URLRequest
	_ = json.NewDecoder(r.Body).Decode(&urlRequest)

	url := fmt.Sprintf("%s://%v", urlRequest.Protocol, urlRequest.URL)

	log.Infof("%+v\n", urlRequest)

	if urlRequest.Fullscreen == 1 {
		log.Info("Entering fullscreen")
		c.fullscreen()
	}
	if urlRequest.HideScrollbar == 1 {
		log.Info("Hiding Scrollbar")
		c.hideScrollbar()
	}
	c.navigate(url)
	if urlRequest.ReloadSeconds > 0 {
		fmt.Fprintf(w, "Reload set to every %d seconds.", urlRequest.ReloadSeconds)
		err := c.reloadInterval(time.Duration(urlRequest.ReloadSeconds) * time.Second)
		if err != nil {
			log.Error(err)
		}
	}
}

// GetFullscreen does not work...
func (c *Chrome) GetFullscreen(w http.ResponseWriter, r *http.Request) {
	log.Info("Entering fullscreen")
	c.fullscreen()
}

// GetReload reloads the browser
func (c *Chrome) GetReload(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Reload")
	c.reload()
}

// NewChrome instantiates a CDP from the given context
func NewChrome(context context.Context) (chrome *Chrome) {
	var discardLog chromedp.LogFunc
	discardLog = func(string, ...interface{}) {}

	option := chromedp.WithLog(discardLog)
	c, err := chromedp.New(context, chromedp.WithLog(log.Printf), option)
	if err != nil {
		log.Fatal(err)
	}
	return &Chrome{Context: &context, CDP: c}
}

func main() {
	context, cancel := context.WithCancel(context.Background())
	defer cancel()
	port := ":8000"
	c := NewChrome(context)
	router := mux.NewRouter()
	router.HandleFunc("/youtube/{id}", c.GetYT).Methods("GET")
	router.HandleFunc("/kibana", c.GetKibana).Methods("GET")
	router.HandleFunc("/url/{protocol}/{url}", c.GetURL).Methods("GET")
	router.HandleFunc("/url", c.PostURL).Methods("POST")
	router.HandleFunc("/fullscreen", c.GetFullscreen).Methods("GET")
	router.HandleFunc("/reload", c.GetReload).Methods("GET")
	log.Infof("Listening on %s", port)
	log.Fatal(http.ListenAndServe(port, router))
}

func cancelReload() {
	if reloadActive {
		log.Debug("Cancel Reload")
		close(cancelChan)
		reloadActive = false
	}
}

func (c *Chrome) navigate(url string) error {

	cancelReload()
	return c.CDP.Run(*c.Context, chromedp.Tasks{chromedp.Navigate(url)})
}

func (c *Chrome) fullscreen() error {
	var buf []byte
	task := chromedp.Tasks{
		chromedp.Evaluate("document.body.webkitRequestFullScreen()", buf),
	}
	return c.CDP.Run(*c.Context, task)
}

func (c *Chrome) hideScrollbar() error {
	var buf []byte
	task := chromedp.Tasks{
		chromedp.Evaluate("document.body.style.overflow = 'hidden'", buf),
	}
	return c.CDP.Run(*c.Context, task)
}

func (c *Chrome) reload() error {
	log.Debug("Reload")
	task := chromedp.Tasks{
		chromedp.Reload(),
	}
	return c.CDP.Run(*c.Context, task)
}

func (c *Chrome) reloadInterval(interval time.Duration) (err error) {
	timeChan := time.Tick(interval)
	cancelChan = make(chan struct{})
	reloadActive = true
	go func() {
		for {
			select {
			case <-cancelChan:
				log.Debug("Cancel Reload")
				return
			case <-timeChan:
				c.reload()
			}
		}
	}()
	return
}
