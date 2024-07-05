package news

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"micro.mu"

	"github.com/mmcdole/gofeed"
)

//go:embed feeds.json
var f embed.FS

var feeds = map[string]string{}

var status = map[string]*Feed{}

// yes I know its hardcoded
var key = os.Getenv("CRYPTO_API_KEY")

// sunnah api key
var sunnah_key = os.Getenv("SUNNAH_API_KEY")

type Feed struct {
	Name     string
	URL      string
	Error    error
	Attempts int
	Backoff  time.Time
}

type Article struct {
	Title       string
	Description string
	URL         string
	Published   string
	Category    string
	PostedAt    time.Time
}

func getPrice(v ...string) map[string]string {
	rsp, err := http.Get(fmt.Sprintf("https://min-api.cryptocompare.com/data/pricemulti?fsyms=%s&tsyms=USD&api_key=%s", strings.Join(v, ","), key))
	if err != nil {
		return nil
	}
	b, _ := ioutil.ReadAll(rsp.Body)
	defer rsp.Body.Close()
	var res map[string]interface{}
	json.Unmarshal(b, &res)
	if res == nil {
		return nil
	}
	prices := map[string]string{}
	for _, t := range v {
		rsp := res[t].(map[string]interface{})
		prices[t] = fmt.Sprintf("%v", rsp["USD"].(float64))
	}
	return prices
}

var tickers = []string{"BTC", "BNB", "ETH", "SOL"}

var replace = []func(string) string{
	func(v string) string {
		return strings.Replace(v, "© 2024 TechCrunch. All rights reserved. For personal use only.", "", -1)
	},
	func(v string) string {
		return regexp.MustCompile(`<img .*>`).ReplaceAllString(v, "")
	},
}

var news = []byte{}
var mutex sync.RWMutex

func addHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		r.ParseForm()
		name := r.Form.Get("name")
		feed := r.Form.Get("feed")
		if len(name) == 0 || len(feed) == 0 {
			http.Error(w, "missing name or feed", 500)
			return
		}

		mutex.Lock()
		_, ok := feeds[name]
		if ok {
			mutex.Unlock()
			http.Error(w, "feed exists with name "+name, 500)
			return
		}

		// save it
		feeds[name] = feed
		mutex.Unlock()

		saveFeed()

		// redirect
		http.Redirect(w, r, "/", 302)
	}

	form := `
<h1>Add Feed</h1>
<form id="add" action="/add" method="post">
<input id="name" name="name" placeholder="feed name" required>
<br><br>
<input id="feed" name="feed" placeholder="feed url" required>
<br><br>
<button>Submit</button>
<p><small>Feed will be parsed in 1 minute</small></p>
</form>
`

	html := mu.Template("Add Feed", "Add a news feed", "", form)

	mu.Render(w, html)
}

func saveFeed() {
	mutex.Lock()
	defer mutex.Unlock()
	file := filepath.Join(mu.Cache, "feeds.json")
	feed, _ := json.Marshal(feeds)
	os.WriteFile(file, feed, 0644)
}

func saveHtml(head, data []byte) {
	if len(data) == 0 {
		return
	}
	html := mu.Template("News", "Read the news", string(head), string(data))
	mutex.Lock()
	news = []byte(html)
	mutex.Unlock()
	cache := filepath.Join(mu.Cache, "news.html")
	os.WriteFile(cache, news, 0644)
}

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	mutex.RLock()
	defer mutex.RUnlock()
	w.Write(news)
}

func StatusHandler(w http.ResponseWriter, r *http.Request) {
	mutex.RLock()
	b, _ := json.Marshal(status)
	mutex.RUnlock()
	w.Header().Set("Content-Type", "application/json")
	w.Write(b)
}

func loadFeed() {
	// load the feeds file
	data, _ := f.ReadFile("feeds.json")
	// unpack into feeds
	mutex.Lock()
	if err := json.Unmarshal(data, &feeds); err != nil {
		fmt.Println("Error parsing feeds.json", err)
	}
	mutex.Unlock()

	// load from cache
	file := filepath.Join(mu.Cache, "feeds.json")

	_, err := os.Stat(file)
	if err == nil {
		// file exists
		b, err := ioutil.ReadFile(file)
		if err == nil && len(b) > 0 {
			var res map[string]string
			json.Unmarshal(b, &res)
			mutex.Lock()
			for name, feed := range res {
				_, ok := feeds[name]
				if ok {
					continue
				}
				fmt.Println("Loading", name, feed)
				feeds[name] = feed
			}
			mutex.Unlock()
		}
	}
}

func parseFeed() {
	cache := filepath.Join(mu.Cache, "news.html")

	f, err := os.Stat(cache)
	if err == nil && len(news) == 0 {
		fmt.Println("Reading cache")
		mutex.Lock()
		news, _ = os.ReadFile(cache)
		mutex.Unlock()

		if time.Since(f.ModTime()) < time.Minute {
			time.Sleep(time.Minute)
		}
	}

	p := gofeed.NewParser()

	data := []byte{}
	head := []byte{}
	urls := map[string]string{}
	stats := map[string]Feed{}

	var sorted []string

	mutex.RLock()
	for name, url := range feeds {
		sorted = append(sorted, name)
		urls[name] = url

		if stat, ok := status[name]; ok {
			stats[name] = *stat
		}
	}
	mutex.RUnlock()

	sort.Strings(sorted)

	var headlines []*Article

	for _, name := range sorted {
		feed := urls[name]

		// check last attempt
		stat, ok := stats[name]
		if !ok {
			stat = Feed{
				Name: name,
				URL:  feed,
			}

			mutex.Lock()
			status[name] = &stat
			mutex.Unlock()
		}

		// it's a reattempt, so we need to check what's going on
		if stat.Attempts > 0 {
			// there is still some time on the clock
			if time.Until(stat.Backoff) > time.Duration(0) {
				// skip this iteration
				continue
			}

			// otherwise we've just hit our threshold
			fmt.Println("Reattempting pull of", feed)
		}

		// parse the feed
		f, err := p.ParseURL(feed)
		if err != nil {
			// up the attempts
			stat.Attempts++
			// set the error
			stat.Error = err
			// set the backoff
			stat.Backoff = time.Now().Add(mu.Backoff(stat.Attempts))
			// print the error
			fmt.Printf("Error parsing %s: %v, attempt %d backoff until %v", feed, err, stat.Attempts, stat.Backoff)

			mutex.Lock()
			status[name] = &stat
			mutex.Unlock()

			// skip ahead
			continue
		}

		mutex.Lock()
		// successful pull
		stat.Attempts = 0
		stat.Backoff = time.Time{}
		stat.Error = nil

		// readd
		status[name] = &stat
		mutex.Unlock()

		head = append(head, []byte(`<a href="#`+name+`" class="head">`+name+`</a>`)...)

		data = append(data, []byte(`<div class=section>`)...)
		data = append(data, []byte(`<hr id="`+name+`" class="anchor">`)...)
		data = append(data, []byte(`<h1>`+name+`</h1>`)...)

		for i, item := range f.Items {
			// only 10 items
			if i >= 10 {
				break
			}

			for _, fn := range replace {
				item.Description = fn(item.Description)
			}

			val := fmt.Sprintf(`
<h3><a href="%s" rel="noopener noreferrer" target="_blank">%s</a></h3>
<span class="description">%s</span>
			`, item.Link, item.Title, item.Description)
			data = append(data, []byte(val)...)

			if i > 0 {
				continue
			}

			headlines = append(headlines, &Article{
				Title:       item.Title,
				Description: item.Description,
				URL:         item.Link,
				Published:   item.Published,
				PostedAt:    *item.PublishedParsed,
				Category:    name,
			})
		}

		data = append(data, []byte(`</div>`)...)
	}

	// head = append(head, []byte(`<a href="/add" class="head"><button>Add</button></a>`)...)

	headline := []byte(`<div class=section><hr id="headlines" class="anchor">`)

	// get hadith
	hadith := getSunnah()
	if len(hadith) > 0 {
		headline = append(headline, []byte(fmt.Sprintf(`<div id="hadith"><h1>Hadith</h1>%s</div>`, hadith))...)
	}

	// get crypto prices
	prices := getPrice(tickers...)

	if prices != nil {
		btc := prices["BTC"]
		eth := prices["ETH"]
		bnb := prices["BNB"]
		sol := prices["SOL"]

		var info []byte
		info = append(info, []byte(`<div id="info"><h1>Markets</h1>`)...)
		info = append(info, []byte(`<span class="ticker">btc $`+btc+`</span>`)...)
		info = append(info, []byte(`<span class="ticker">eth $`+eth+`</span>`)...)
		info = append(info, []byte(`<span class="ticker">bnb $`+bnb+`</span>`)...)
		info = append(info, []byte(`<span class="ticker">sol $`+sol+`</span>`)...)
		info = append(info, []byte(`</div>`)...)
		headline = append(headline, info...)
	}

	headline = append(headline, []byte(`<h1>Headlines</h1>`)...)

	// create the headlines
	sort.Slice(headlines, func(i, j int) bool {
		return headlines[i].PostedAt.After(headlines[j].PostedAt)
	})

	for _, h := range headlines {
		val := fmt.Sprintf(`
			<div class="headline"><a href="#%s" class="category">%s</a><h3><a href="%s" rel="noopener noreferrer" target="_blank">%s</a></h3><span class="description">%s</span></div>`,
			h.Category, h.Category, h.URL, h.Title, h.Description)
		headline = append(headline, []byte(val)...)
	}

	headline = append(headline, []byte(`</div>`)...)

	// set the headline
	data = append(headline, data...)

	// save it
	saveHtml(head, data)

	// wait 10 minutes
	time.Sleep(time.Minute * 10)

	// go again
	parseFeed()
}

func getSunnah() string {
	if len(sunnah_key) == 0 {
		return ""
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	books := map[string]int{
		"bukhari": 7563,
		"muslim":  3033,
	}

	var hadiths []string

	for book, limit := range books {
		for i := 0; i < 3; i++ {
			hadith := r.Intn(limit)

			// reset
			if hadith == 0 {
				hadith = 1
			}

			uri := fmt.Sprintf("https://api.sunnah.com/v1/collections/%s/hadiths/%d", book, hadith)
			req, err := http.NewRequest("GET", uri, nil)
			if err != nil {
				return ""
			}

			req.Header.Set("X-API-Key", sunnah_key)
			req.Header.Set("Accept", "application/json")

			rsp, err := http.DefaultClient.Do(req)
			if err != nil {
				continue
			}
			defer rsp.Body.Close()
			b, _ := ioutil.ReadAll(rsp.Body)

			var resp map[string]interface{}
			json.Unmarshal(b, &resp)

			if v := resp["error"]; v != nil {
				continue
			}

			h := resp["hadith"].([]interface{})
			if len(h) == 0 {
				continue
			}
			had := h[0].(map[string]interface{})
			title := had["chapterTitle"].(string)
			text := had["body"].(string)

			hadiths = append(hadiths, fmt.Sprintf(`<div><b>%s</b><br>%s<a href="https://sunnah.com/%s:%d">%s:%d</a></div>`, title, text, book, hadith, book, hadith))
			break
		}
	}

	return strings.Join(hadiths, "<br>")
}

func FeedsHandler(w http.ResponseWriter, r *http.Request) {
	mutex.RLock()
	defer mutex.RUnlock()

	data := `<h1>Feeds</h1>`

	for name, feed := range feeds {
		data += fmt.Sprintf(`<a href="%s">%s</a><br>`, feed, name)
	}

	html := mu.Template("Feeds", "News RSS feeds", "", data)
	mu.Render(w, html)
}

func Register() {
	// load the feeds
	loadFeed()

	go parseFeed()
}
