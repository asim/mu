package chat

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	//"net/url"
	"sort"
	//"strings"
	"sync"

	"github.com/micro/mu"

	"github.com/google/uuid"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"

	"github.com/sashabaranov/go-openai"
)

type Channel struct {
	Name     string
	Topic    string
	Messages []string
}

var channels = map[string]*Channel{
	"general": new(Channel),
	"finance": new(Channel),
	"health":  new(Channel),
	"islam":   new(Channel),
	"misc":    new(Channel),
}

type Command func(*Channel, string) string

var commands = map[string]Command{
	"openai": func(channel *Channel, prompt string) string {
		key := os.Getenv("OPENAI_API_KEY")
		if len(key) == 0 {
			return ""
		}

		client := openai.NewClient(key)

		var message []openai.ChatCompletionMessage

		var tokens int

		for i := len(channel.Messages); i > 0; i-- {
			msg := channel.Messages[i-1]

			// TODO: fix role
			message = append([]openai.ChatCompletionMessage{{
				Role:    openai.ChatMessageRoleUser,
				Content: msg,
			}}, message...)

			tokens += len(msg)

			if tokens > 4096 {
				break
			}
		}

		message = append(message, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleUser,
			Content: prompt,
		})

		req := openai.ChatCompletionRequest{
			Model:     openai.GPT3Dot5Turbo,
			Messages:  message,
			User:      channel.Name,
			MaxTokens: 4096,
		}

		resp, err := client.CreateChatCompletion(context.Background(), req)

		var reply string
		if err != nil {
			reply = err.Error()
		} else {
			reply = resp.Choices[0].Message.Content
		}

		return reply
	},
}

var updates = make(chan bool, 1)

var mutex sync.RWMutex

func mdToHTML(md []byte) []byte {
	// create markdown parser with extensions
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs | parser.NoEmptyLineBeforeBlock
	p := parser.NewWithExtensions(extensions)
	doc := p.Parse(md)

	// create HTML renderer with extensions
	htmlFlags := html.CommonFlags | html.HrefTargetBlank
	opts := html.RendererOptions{Flags: htmlFlags}
	renderer := html.NewRenderer(opts)

	return markdown.Render(doc, renderer)
}

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	id := uuid.New().String()
	channel := "general"

	// get cookie
	c, err := r.Cookie("uuid")
	if err == nil && len(c.Value) > 0 {
		id = c.Value
	} else {
		http.SetCookie(w, &http.Cookie{
			Name:  "uuid",
			Value: id,
		})
	}

	c, err = r.Cookie("channel")
	if err == nil && len(c.Value) > 0 {
		channel = c.Value
	} else {
		http.SetCookie(w, &http.Cookie{
			Name:  "channel",
			Value: channel,
		})
	}

	mutex.Lock()
	ch, ok := channels[channel]
	if !ok {
		ch = new(Channel)
		channels[channel] = ch
	}
	mutex.Unlock()

	// get the channel
	text := ""
	for i, m := range ch.Messages {
		class := "message"

		mod := i % 2
		if mod != 0 {
			class = "message mu"
		}

		text += fmt.Sprintf(`<div class="%s">%s</div>`, class, m)
	}

	t := mu.Template("Chat", "Ask an AI", `
      <a href="#general" class="head">General</a>
      <a href="#finance" class="head">Finance</a>
      <a href="#health" class="head">Health</a>
      <a href="#islam" class="head">Islam</a>
      <a href="#misc" class="head">Misc</a>`, `
    <style>
      #input {
	width: 100%;
	height: 55px;
	position: relative;
      }
       #prompt {
         width: calc(100% - 100px);
	 padding: 10px;
       }
       #form {
	 bottom: 10px;
	 margin: 5px;
       }
       #form button {
         padding: 10px;
       }
       .message {
         padding: 10px 10px;
       }
       #text {
	 height: calc(100% - 140px);
	 overflow-y: scroll;
	 padding-top: 50px;
       }
       .highlight {
         text-decoration: underline;
       }
       .mu {
         background: #F8F8F8;
	 border-radius: 10px;
       }
       @media only screen and (max-width: 600px) {
         #text { padding: 40px 0 20px 0; }
       }
    </style>

    <div id=text>`+text+`</div>

    <div id="input">
      <form id="form" action="/prompt">
        <input id="uuid" name="uuid" type="hidden" value="`+id+`">
        <input id="prompt" name="prompt" placeholder="ask a question" autocomplete="off">
	<input id="channel" name="channel" type="hidden" value="`+channel+`">
        <button>submit</button>
      </form>
    </div>

    <script>
	String.prototype.parseURL = function(embed) {
	    return this.replace(/[A-Za-z]+:\/\/[A-Za-z0-9-_]+\.[A-Za-z0-9-_:%&~\?\/.=]+/g, function(url) {
		if (embed == true) {
		    var match = url.match(/^.*(youtu.be\/|v\/|u\/\w\/|embed\/|watch\?v=|\&v=)([^#\&\?]*).*/);
		    if (match && match[2].length == 11) {
			return '<div class="iframe">' +
			    '<iframe src="//www.youtube.com/embed/' + match[2] +
			    '" frameborder="0" allowfullscreen></iframe>' + '</div>';
		    };
		    if (url.match(/^.*giphy.com\/media\/[a-zA-Z0-9]+\/[a-zA-Z0-9]+.gif$/)) {
			return '<div class="animation"><img src="' + url + '"></div>';
		    }
		};
		// var pretty = url.replace(/^http(s)?:\/\/(www\.)?/, '');
		//return pretty.link(url);
		return url.link(url)
	    });
	};

      var form = document.getElementById("form");
      var text = document.getElementById("text");

      // parse and embed
      text.innerHTML = text.innerHTML.parseURL();

      form.addEventListener("submit", function(ev) {
	ev.preventDefault();
        var data = document.getElementById("form");
	var uuid = form.elements["uuid"].value;
        var prompt = form.elements["prompt"].value;
	var channel = form.elements["channel"].value;
	form.elements["prompt"].value = '';
	text.innerHTML += "<div class='message mu'>" + prompt.parseURL() + "</div>";
	text.scrollTo(0, text.scrollHeight);
	var data = {"uuid": uuid, "prompt": prompt, "markdown": true, channel: channel};

	fetch("/chat/prompt", {
		method: "POST",
		body: JSON.stringify(data),
		headers: {'Content-Type': 'application/json'},
	})
	  .then(res => res.json())
	  .then((rsp) => {
		  if (rsp.answer === undefined) {
			return
		  }
		  if (rsp.markdown === undefined) {
			return
		  }
		  var answer = rsp.markdown;
		  var height = text.scrollHeight;
		  text.innerHTML += "<div class=message>" + answer + "</div>";
		  text.scrollTo(0, height + 20);
	});
	return false;
      });

      var hash = window.location.hash.replace("#", "");

      if (hash.length == 0) {
        hash = "`+channel+`";
      }

      var el = document.querySelectorAll('#nav a');
      for (let i = 0; i < el.length; i++) {
	if (el[i].href == "/") {
	  continue;
	}
        el[i].className = 'head';
        if (el[i].href.endsWith('#' + hash)) {
          el[i].className = 'highlight head';
        }
      }

      document.cookie = "channel="+"`+channel+`";

      text.scrollTo(0, text.scrollHeight);

      window.addEventListener("hashchange", () => {
        var hash = window.location.hash.replace("#", "");
	var channel = document.getElementById("channel")
	channel.value = hash;
        document.cookie = "channel="+hash;

	window.location.reload();
      }, false);

      window.onload = (event) => { 
        
      }
    </script>
	`)
	mu.Render(w, t)
}

type Req struct {
	UUID     string `json:"uuid"`
	Prompt   string `json:"prompt"`
	Markdown bool   `json:"markdown",omitempty`
	Channel  string `json:"channel",omitempty`
}

func ChannelHandler(w http.ResponseWriter, r *http.Request) {
	mutex.Lock()

	html := "<h1>Channels</h1>"

	var chans []string

	for ch, _ := range channels {
		chans = append(chans, ch)
	}

	sort.Strings(chans)

	for _, ch := range chans {
		if len(ch) == 0 {
			continue
		}
		html += fmt.Sprintf(`<a href="/#%s">%s</a><br>`, ch, ch)
	}
	mutex.Unlock()

	t := mu.Template("Channels", "List of channels", "", html)
	mu.Render(w, t)
}

func PromptHandler(w http.ResponseWriter, r *http.Request) {
	b, _ := ioutil.ReadAll(r.Body)
	var req Req
	json.Unmarshal(b, &req)

	id := req.UUID
	prompt := req.Prompt

	if len(req.UUID) == 0 {
		fmt.Println("uuid", id)
		return
	}
	if len(req.Prompt) == 0 {
		fmt.Println("no prompt")
		return
	}

	if len(req.Channel) == 0 {
		req.Channel = "general"
	}

	mutex.Lock()
	c, ok := channels[req.Channel]
	if ok {
		c.Messages = append(c.Messages, prompt)
	}
	mutex.Unlock()

	select {
	case updates <- true:
	default:
	}

	// ask the question
	//parts := strings.Split(prompt, " ")
	//if parts[0] != "/ai" {
	//	w.Write([]byte(`{}`))
	//	return
	//}

	command, ok := commands["openai"]
	if ok {
		//answer := command(c, strings.Join(parts[1:], " "))
		answer := command(c, prompt)
		markdown := ""
		message := answer

		if req.Markdown {
			markdown = string(mdToHTML([]byte(answer)))
			message = markdown
		}

		// get the answer
		rsp := map[string]interface{}{
			"answer":   answer,
			"markdown": markdown,
		}

		mutex.Lock()
		c, ok := channels[req.Channel]
		if ok {
			c.Messages = append(c.Messages, message)
		}
		mutex.Unlock()

		select {
		case updates <- true:
		default:
		}

		b, _ = json.Marshal(rsp)
		w.Header().Set("Content-Type", "application/json")
		w.Write(b)
		return
	}

	w.Write([]byte(`{}`))
}

func load() {
	mutex.Lock()
	mu.Load(&channels, "chat.enc", true)
	mutex.Unlock()
}

func save() {
	for {
		select {
		case <-updates:
			mutex.RLock()
			mu.Save(channels, "chat.enc", true)
			mutex.RUnlock()
		}
	}
}

func Register() {
	load()

	go save()
}
