package home

import (
	"fmt"
	"net/http"

	"github.com/micro/mu"
)

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	nav := `<a href="/logout" class="head">Logout</a>`

	c, err := r.Cookie("user")
	if err == nil && len(c.Value) > 0 {
		nav = fmt.Sprintf("<a href=# class=head><b>%s</b></a>", c.Value) + nav
	}

	html := mu.Template("Home", "Home screen", nav, `
<style>
#title {
  margin-top: 100px;
}
.apps {
  max-width: 600px;
}
.apps a {
  margin-right: 10px;
}
</style>
          <h1 id="title">Home</h1>
          <p id="description"></p>

	  <div class="apps">
	    <a href="/chat">
	      <button>
		Chat
	      </button>
	    </a>
	    <a href="/news">
	      <button>
		News
	      </button>
	    </a>
	    <a href="/pray">
	      <button>
		Pray
	      </button>
	    </a>
	    <a href="/reminder">
	      <button>
		Reminder
	      </button>
	    </a>
	    <a href="/watch">
	      <button>
		Watch
	      </button>
	    </a>
	    <a href="/work">
	      <button>
		Work
	      </button>
	    </a>
	  </div>
	`)

	mu.Render(w, html)
}

func Register() {}
