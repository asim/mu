package reminder

import (
	"embed"
	"encoding/json"
	"fmt"
	"micro.mu"
	"net/http"
)

//go:embed quran/*
var quran embed.FS

var Quran map[string]string

var Source = "https://api.alquran.cloud/v1/quran/en.sahih"

var HTML string

func load() {
	if err := mu.Load(&HTML, "quran.html", false); err == nil {
		return
	}

	if err := mu.Load(&Quran, "quran.dev", false); err == nil {
		HTML = html()
		mu.Save(HTML, "quran.html", false)
		return
	}

	// get the quran
	fmt.Println("Loading source")

	Quran = make(map[string]string)

	// Set local
	for i := 0; i < 114; i++ {
		f, err := quran.ReadFile(fmt.Sprintf("quran/%d.json", i+1))
		if err != nil {
			panic(err.Error())
		}
		var data []interface{}
		json.Unmarshal(f, &data)

		name := data[0].(map[string]interface{})["name"].(map[string]interface{})["transliterated"].(string)
		name += "<br>"
		name += data[0].(map[string]interface{})["name"].(map[string]interface{})["translated"].(string)

		data = data[1:]

		// set the name
		Quran[fmt.Sprintf("%d", i)] = name

		for j, ayah := range data {
			key := fmt.Sprintf("%d:%d", i, j)
			// save the text for the ayah
			Quran[key] = fmt.Sprintf("%v", ayah.([]interface{})[1].(string))
		}
	}

	fmt.Println("Compiling Quran")

	fmt.Println("Saving to cache")

	// save it it
	if err := mu.Save(Quran, "quran.dev", false); err != nil {
		panic(err.Error())
	}

	// save html
	HTML = html()
	mu.Save(HTML, "quran.html", false)
}

var html = func() string {
	var data string

	data = `<style>
  .ayah {
    padding: 5px 0 5px 0;
    max-width: 600px;
  }
  .ayah a:hover {
    background: whitesmoke;
  }
  .marker {
    padding-top: 5px;
    display: block;
  }
  .marker a {
    font-size: 0.8em;
    color: grey;
  }
  .surah {
    padding: 50px 0 100px 0;
  }
  .surah h1 {
    font-size: 3.5em;
  }
</style>
<div style="margin-top: 100px">
<form id=goto>
  <input id="marker" placeholder=114>
  <button>goto</button>
</form>
</div>
<script>
function goto(ev) {
  ev.preventDefault()
  var marker = document.getElementById("marker");
  var val = marker.value;
  console.log(val);
  window.location.hash = "#" + val;

  return false;
}

var form = document.getElementById("goto");
form.onsubmit = goto;

document.addEventListener("scroll", (event) => {
  console.log("setting marker", window.scrollY);
  localStorage.setItem("marker", window.scrollY);
});

document.addEventListener('DOMContentLoaded', function() {
  var pos = localStorage.getItem("marker");
  if (pos == undefined) {
	  return
  }
  window.scrollTo(0, pos);
}, false);
</script>
`

	// 114 surahs
	for i := 0; i < 114; i++ {
		name := Quran[fmt.Sprintf("%d", i)]

		data += fmt.Sprintf(`<div id="%d" class="surah"><h1>%d</h1><p>%s</p>`, i+1, i+1, name)

		// max 286 ayahs
		for j := 0; j < 286; j++ {
			key := fmt.Sprintf("%d:%d", i, j)
			text, ok := Quran[key]
			if !ok {
				break
			}
			ref := fmt.Sprintf("%d:%d", i+1, j+1)
			link := fmt.Sprintf(`<a href="https://quran.com/%s">%s</a>`, ref, text)
			data += fmt.Sprintf(`<div id=%s class="marker"><a href="#%s">%s</a></div>`, ref, ref, ref)
			data += fmt.Sprintf("<div class=ayah>%s</div>", link)
		}

		data += "</div>"
	}

	return data
}

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	nav := `
	<a href="#1" class=head>The Opening</a>
	<a href="#2:255" class=head>The Throne</a>
	<a href="#112" class=head>Sincerity</a>
	<a href="#113" class=head>The Dawn</a>
	<a href="#114" class=head>Mankind</a>
	`

	html := mu.Template("Reminder", "Read the Quran", nav, HTML)
	mu.Render(w, html)
}

func Register() {
	load()
}
