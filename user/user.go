// Package user is for authentication
package user

import (
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"os"
	"sort"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"

	"micro.mu"
)

var mutex sync.Mutex
var users = map[string]*Account{}
var sessions = map[string]*Session{}

var admin = os.Getenv("USER_ADMIN")

func init() {
	// load users
	mu.Load(&users, "users.enc", true)
	mu.Load(&sessions, "sessions.enc", true)
}

type Account struct {
	ID       string
	Username string
	Password string
	Created  time.Time
}

type Session struct {
	ID       string
	Username string
}

func newSess(acc *Account) *Session {
	id := mu.ID()
	sess := &Session{
		ID:       base64.StdEncoding.EncodeToString([]byte(id)),
		Username: acc.Username,
	}
	return sess
}

func Get(r *http.Request) (*Account, error) {
	c, err := r.Cookie("user")
	if err != nil {
		return nil, err
	}
	if len(c.Value) == 0 {
		return nil, errors.New("no user")
	}
	acc, ok := users[c.Value]
	if !ok {
		return nil, errors.New("no user")
	}
	return acc, nil
}

// Admin is the user admin
func Admin(w http.ResponseWriter, r *http.Request) {
	// get user cookie
	c, err := r.Cookie("user")
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	if len(c.Value) == 0 || len(admin) == 0 {
		return
	}

	// check if its the admin
	if c.Value != admin {
		return
	}

	var div string

	var userList []string
	mutex.Lock()
	for _, user := range users {
		userList = append(userList, user.Username)
	}
	mutex.Unlock()
	sort.Strings(userList)

	for _, user := range userList {
		div += fmt.Sprintf(`<div class="user">%s</div>`, user)
	}

	// list the users
	html := mu.Template("Admin", "User Admin", "", `<h1 style="padding-top: 100px;">Users</h1>`+div)
	mu.Render(w, html)
}

func Exists(u string) bool {
	mutex.Lock()
	defer mutex.Unlock()

	_, ok := users[u]
	return ok
}

// Login a user
func Login(username, password string) (*Account, *Session, error) {
	mutex.Lock()
	defer mutex.Unlock()

	acc, ok := users[username]
	if !ok {
		return nil, nil, errors.New("does not exist")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(acc.Password), []byte(password)); err != nil {
		return nil, nil, errors.New("invalid password")
	}

	sess := newSess(acc)
	sessions[sess.ID] = sess

	mu.Save(sessions, "sessions.enc", true)

	return acc, sess, nil
}

// Logout the user
func Logout(username string, sess *Session) error {
	mutex.Lock()
	defer mutex.Unlock()

	// logout
	sessions[sess.ID] = nil

	// TODO: ...
	return mu.Save(sessions, "sessions.enc", true)
}

// Signup a user
func Signup(username, password string) error {
	if _, ok := users[username]; ok {
		return errors.New("already exists")
	}

	pw, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	mutex.Lock()
	users[username] = &Account{
		ID:       mu.ID(),
		Username: username,
		Password: string(pw),
	}
	mutex.Unlock()

	// save accounts
	return mu.Save(users, "users.enc", true)
}

// Verify a session
func Verify(sessID, user string) error {
	mutex.Lock()
	defer mutex.Unlock()

	sess, ok := sessions[sessID]
	if !ok {
		return errors.New("invalid session")
	}
	if sess == nil {
		return errors.New("expired session")
	}
	if user != sess.Username {
		return errors.New("invalid user")
	}

	return nil
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		r.ParseForm()
		user := r.Form.Get("username")
		pass := r.Form.Get("password")

		acc, sess, err := Login(user, pass)
		if err != nil {
			http.Error(w, err.Error(), 401)
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:    "sess",
			Value:   sess.ID,
			Expires: time.Now().Add(time.Hour * 24 * 30),
		})

		http.SetCookie(w, &http.Cookie{
			Name:    "user",
			Value:   acc.Username,
			Expires: time.Now().Add(time.Hour * 24 * 30),
		})

		http.Redirect(w, r, "/home", 302)
		return
	}

	// Login screen
	html := mu.Template("Login", "Login to your account", "", `
<style>
  #login {
    padding-top: 100px;
  }
</style>
<div id="login">
<h1>Login</h1>
<form action="/login", method="post">
  <input id="username" name="username" placeholder=Username>
  <br><br>
  <input id="password" name="password" type="password" placeholder=Password>
  <br><br>
  <button>Submit</button>
</form>
</div>
`)
	mu.Render(w, html)
}

func SignupHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		r.ParseForm()
		user := r.Form.Get("username")
		pass := r.Form.Get("password")

		err := Signup(user, pass)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		acc, sess, err := Login(user, pass)
		if err != nil {
			http.Error(w, err.Error(), 401)
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:    "sess",
			Value:   sess.ID,
			Expires: time.Now().Add(time.Hour * 24 * 30),
		})

		http.SetCookie(w, &http.Cookie{
			Name:    "user",
			Value:   acc.Username,
			Expires: time.Now().Add(time.Hour * 24 * 30),
		})

		http.Redirect(w, r, "/home", 302)
		return
	}

	// Login screen
	html := mu.Template("Signup", "Signup for an account", "", `
<style>
  #signup {
    padding-top: 100px;
  }
</style>
<div id="signup">
<h1>Signup</h1>
<form action="/signup", method="post">
  <input id="username" name="username" placeholder=Username>
  <br><br>
  <input id="password" name="password" type="password" placeholder=Password>
  <br><br>
  <button>Submit</button>
</form>
</div>
`)
	mu.Render(w, html)
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:   "sess",
		MaxAge: -1,
	})
	http.SetCookie(w, &http.Cookie{
		Name:   "user",
		MaxAge: -1,
	})
	http.Redirect(w, r, "/", 302)
}

func Register() {}

// Authenticated handler
func Auth(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// check the session cookie exists
		c, err := r.Cookie("sess")
		if err != nil || len(c.Value) == 0 {
			http.Redirect(w, r, "/login", 302)
			return
		}
		// check the session cookie exists
		cu, err := r.Cookie("user")
		if err != nil || len(cu.Value) == 0 {
			http.Redirect(w, r, "/login", 302)
			return
		}

		// TODO: check the cookie is valid
		if err := Verify(c.Value, cu.Value); err != nil {
			http.Redirect(w, r, "/login", 302)
			return
		}

		// run the handler
		h(w, r)
	}
}
