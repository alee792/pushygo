package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/gorilla/mux"
)

var c *chromedp.CDP

type Chrome struct {
	Context *context.Context
	CDP     *chromedp.CDP
}

type YT struct {
	ID string `json:"id,omitempty"`
}

func (c *Chrome) GetYT(w http.ResponseWriter, r *http.Request) {

	fmt.Fprintf(w, "Getting Youtube.")
	params := mux.Vars(r)
	var yt YT
	_ = json.NewDecoder(r.Body).Decode(&yt)
	yt.ID = params["id"]

	url := fmt.Sprintf("https://youtube.com/embed/%s?autoplay=1", yt.ID)
	err := c.CDP.Run(*c.Context, getYT(url))
	if err != nil {
		log.Fatal(err)
	}
}

func (c *Chrome) GetKibana(w http.ResponseWriter, r *http.Request) {

	fmt.Fprintf(w, "Getting Kibana.")
	url := "http://kibana.middleearth.eltoro.com:5601/app/kibana#/discover?_g=(refreshInterval:('$$hashKey':'object:685',display:'10%20seconds',pause:!f,section:1,value:10000),time:(from:now-24h,mode:quick,to:now))&_a=(columns:!(_source),index:'2fe3dd00-011a-11e8-9ccb-99d5d0a716fa',interval:auto,query:(language:lucene,query:''),sort:!('@timestamp',desc))"

	err := c.CDP.Run(*c.Context, getYT(url))
	if err != nil {
		log.Fatal(err)
	}
}

func (c *Chrome) GetURL(w http.ResponseWriter, r *http.Request) {

	params := mux.Vars(r)
	fmt.Fprintf(w, "Getting URL %s.", params["url"])
	https := params["https"]
	if https == "" {
		https = "https"
	}
	url := fmt.Sprintf("%s://%v", params["https"], params["url"])
	err := c.CDP.Run(*c.Context, getYT(url))
	if err != nil {
		log.Fatal(err)
	}
}

func NewChrome(context context.Context) (chrome *Chrome) {
	c, err := chromedp.New(context, chromedp.WithLog(log.Printf))
	if err != nil {
		log.Fatal(err)
	}
	return &Chrome{Context: &context, CDP: c}
}

func main() {
	context, cancel := context.WithCancel(context.Background())
	defer cancel()
	c := NewChrome(context)
	router := mux.NewRouter()
	router.HandleFunc("/youtube/{id}", c.GetYT).Methods("GET")
	router.HandleFunc("/kibana", c.GetKibana).Methods("GET")
	router.HandleFunc("/url/{https}/{url}", c.GetURL).Methods("GET")

	router.HandleFunc("/queue/{url}", c.GetURL).Methods("GET")
	log.Fatal(http.ListenAndServe(":8000", router))
}

func getYT(url string) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.Navigate(url),
		chromedp.Sleep(5 * time.Minute),
	}
}

func Navigate(url string) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.Navigate(url),
		chromedp.Sleep(5 * time.Minute),
	}
}
