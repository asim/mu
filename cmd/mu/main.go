package main

import (
	"net/http"

	"micro.mu"
	"micro.mu/chat"
	"micro.mu/home"
	"micro.mu/news"
	"micro.mu/pray"
	"micro.mu/reminder"
	"micro.mu/user"
	"micro.mu/watch"
)

func main() {
	http.HandleFunc("/robots.txt", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`User-agent: *
Allow: /`))
	})

	// chat
	http.HandleFunc("/chat", chat.IndexHandler)
	http.HandleFunc("/chat/prompt", user.Auth(chat.PromptHandler))
	http.HandleFunc("/chat/channels", user.Auth(chat.ChannelHandler))

	// home
	http.HandleFunc("/home", user.Auth(home.IndexHandler))

	// news
	http.HandleFunc("/news", news.IndexHandler)
	http.HandleFunc("/news/feeds", user.Auth(news.FeedsHandler))
	http.HandleFunc("/news/status", user.Auth(news.StatusHandler))
	// http.HandleFunc("/add", addHandler)

	// pray
	http.HandleFunc("/pray", pray.IndexHandler)

	// reminder
	http.HandleFunc("/reminder", reminder.IndexHandler)

	// user auth
	http.HandleFunc("/admin", user.Auth(user.Admin))
	http.HandleFunc("/login", user.LoginHandler)
	http.HandleFunc("/logout", user.LogoutHandler)
	http.HandleFunc("/signup", user.SignupHandler)

	// watch
	http.HandleFunc("/watch", watch.WatchHandler)

	// any other stuff
	chat.Register()
	home.Register()
	news.Register()
	pray.Register()
	reminder.Register()
	user.Register()
	watch.Register()

	mu.Serve(8080)
}
