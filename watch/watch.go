package watch

import (
	"context"
	"fmt"
	"github.com/micro/mu"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

var Key = os.Getenv("YOUTUBE_API_KEY")
var Client, _ = youtube.NewService(context.TODO(), option.WithAPIKey(Key))

var mutex sync.Mutex

// recent query cache keyed by the query string
var Recent = map[string]string{}

// searches by user
var Searches = map[string][]string{}

func init() {
	mu.Load(&Searches, "searches.enc", true)
	mu.Load(&Recent, "recent.json", false)
}

func embedVideo(id string) string {
	u := "https://www.youtube.com/embed/" + id
	style := `style="position: absolute; top: 0; left: 0; right: 0; width: 100%; height: 100%; border: none;"`
	return `<iframe width="560" height="315" ` + style + ` src="` + u + `" title="YouTube video player" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture" allowfullscreen></iframe>`
}

func getResults(q string) (string, error) {
	resp, err := Client.Search.List([]string{"id", "snippet"}).Q(q).MaxResults(25).Do()
	if err != nil {
		return "", err
	}

	var results string

	for _, item := range resp.Items {
		var id, url, desc string
		kind := strings.Split(item.Id.Kind, "#")[1]
		t, _ := time.Parse(time.RFC3339, item.Snippet.PublishedAt)
		desc = fmt.Sprintf(`[%s] published on %s`, kind, t.Format(time.RFC822))
		switch kind {
		case "video":
			id = item.Id.VideoId
			url = "/watch?id=" + id
		case "playlist":
			id = item.Id.PlaylistId
			url = "https://youtube.com/playlist?list=" + id
		case "channel":
			id = item.Id.ChannelId
			url = "https://www.youtube.com/channel/" + id
			desc = "[channel]"
		}
		channel := fmt.Sprintf(`<a href="https://youtube.com/channel/%s">%s</a>`, item.Snippet.ChannelId, item.Snippet.ChannelTitle)
		results += fmt.Sprintf(`
			<div class="thumbnail"><a href="%s"><img src="%s"><h3>%s</h3></a>%s | %s</div>`,
			url, item.Snippet.Thumbnails.Medium.Url, item.Snippet.Title, channel, desc)
	}

	return results, nil
}

func makeNav(uid string) string {
	// build the nav
	var nav string

	mutex.Lock()

	searches := Searches[uid]

	for i := len(searches); i > 0; i-- {
		k := searches[i-1]
		nav += fmt.Sprintf(`<a class="head" href="/watch?q=%s">%s</a>`, url.QueryEscape(k), k)
	}

	mutex.Unlock()

	return nav
}

func saveSearch(uid string, q string) {
	searches := Searches[uid]

	var seen bool
	for _, k := range searches {
		if q == k {
			seen = true
		}
	}
	if !seen {
		searches = append(searches, q)
	}

	Searches[uid] = searches
	mu.Save(Searches, "searches.enc", true)
}

var Results = `
<style>
  form {
    margin-top: 100px;
  }
  .thumbnail {
    margin-bottom: 50px;
  }
  img {
    border-radius: 10px;
  }
  h3 {
    margin-bottom: 5px;
  }
</style>
<form action="/watch" method="POST">
  <input name="q" id="q" value="%s">
  <button>Submit</button>
</form>
<h1>Results</h1>
<div id="results">
%s
</div>`

var Template = `
<style>
form {
  margin-top: 100px;
}
.video {
  width: 100%;
  overflow: hidden;
  padding-top: 56.25%;
  position: relative;
}
</style>
<form action="/watch" method="POST">
  <input name="q" id="q" placeholder=Search>
  <button>Submit</button>
</form>`

func WatchHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	q := r.Form.Get("q")

	var uid string
	c, err := r.Cookie("user")
	if err != nil {
		uid = "default"
	} else if len(c.Value) > 0 {
		uid = c.Value
	}

	if r.Method == "POST" {
		// check recent cache
		mutex.Lock()
		results, ok := Recent[q]
		saveSearch(uid, q)
		mutex.Unlock()

		if !ok {
			var err error

			// fetch results from api
			results, err = getResults(q)
			if err != nil {
				http.Error(w, err.Error(), 500)
				return
			}

			// recent queries
			mutex.Lock()
			Recent[q] = results
			mu.Save(Recent, "recent.json", false)
			mutex.Unlock()
		}

		nav := makeNav(uid)

		html := mu.Template("Watch", q+" | Results", nav, fmt.Sprintf(Results, q, results))
		mu.Render(w, html)
		return
	}

	id := r.Form.Get("id")
	nav := makeNav(uid)

	// render watch page
	if len(id) > 0 {
		// get the page
		html := fmt.Sprintf(`<html><body><div class="video" style="padding-top: 100px">%s</div></body></html>`, embedVideo(id))
		mu.Render(w, html)
		return
	}

	// GET
	// check recent cache

	if len(q) == 0 {
		html := mu.Template("Watch", "Watch YouTube Videos", nav, Template)
		mu.Render(w, html)
		return
	}

	mutex.Lock()
	results, ok := Recent[q]
	mutex.Unlock()

	var content string

	if ok {
		content = fmt.Sprintf(Results, q, results)
	} else {
		content = Template
	}

	html := mu.Template("Watch", "Watch YouTube Videos", nav, content)

	mu.Render(w, html)
}

func Register() {}
