package mu

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/google/uuid"
)

//go:embed html/*
var html embed.FS

// user home dir
var Home string

// the file cache
var Cache string

// the secret key
var Key string

func init() {
	user, err := os.UserHomeDir()
	if err != nil {
		panic(err.Error())
	}

	home := filepath.Join(user, "mu")
	if err := os.MkdirAll(home, 0700); err != nil {
		panic(err.Error())
	}

	// set home
	Home = home

	cache := filepath.Join(home, "cache")
	if err := os.MkdirAll(cache, 0700); err != nil {
		panic(err.Error())
	}

	// set cache
	Cache = cache

	path := filepath.Join(home, "key")
	b, _ := os.ReadFile(path)

	if len(b) == 0 {
		// generate a new key
		bytes := make([]byte, 32) //generate a random 32 byte key for AES-256
		if _, err := rand.Read(bytes); err != nil {
			panic(err.Error())
		}

		key := hex.EncodeToString(bytes) //encode key in bytes to string and keep as secret, put in a vault
		fmt.Println("generating new key", path)

		// write the file
		if err := os.WriteFile(path, []byte(key), 0600); err != nil {
			panic(err.Error())
		}

		Key = key
	} else {
		fmt.Println("loading key", path)
		Key = string(b)
	}
}

func encrypt(stringToEncrypt string, keyString string) (encryptedString string) {
	//Since the key is in string, we need to convert decode it to bytes
	key, _ := hex.DecodeString(keyString)
	plaintext := []byte(stringToEncrypt)

	//Create a new Cipher Block from the key
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err.Error())
	}

	//Create a new GCM - https://en.wikipedia.org/wiki/Galois/Counter_Mode
	//https://golang.org/pkg/crypto/cipher/#NewGCM
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		panic(err.Error())
	}

	//Create a nonce. Nonce should be from GCM
	nonce := make([]byte, aesGCM.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		panic(err.Error())
	}

	//Encrypt the data using aesGCM.Seal
	//Since we don't want to save the nonce somewhere else in this case, we add it as a prefix to the encrypted data. The first nonce argument in Seal is the prefix.
	ciphertext := aesGCM.Seal(nonce, nonce, plaintext, nil)
	return fmt.Sprintf("%x", ciphertext)
}

func decrypt(encryptedString string, keyString string) (decryptedString string) {
	key, _ := hex.DecodeString(keyString)
	enc, _ := hex.DecodeString(encryptedString)

	//Create a new Cipher Block from the key
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err.Error())
	}

	//Create a new GCM
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		panic(err.Error())
	}

	//Get the nonce size
	nonceSize := aesGCM.NonceSize()

	//Extract the nonce from the encrypted data
	nonce, ciphertext := enc[:nonceSize], enc[nonceSize:]

	//Decrypt the data
	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		panic(err.Error())
	}

	return fmt.Sprintf("%s", plaintext)
}

func save(val interface{}, file, key string, encrypted bool) error {
	cache := filepath.Join(Cache, file)

	// encode data
	data, err := json.Marshal(val)
	if err != nil {
		return err
	}

	// encrypt it
	if encrypted {
		enc := encrypt(string(data), Key)
		data = []byte(enc)
	}

	// write the data
	return os.WriteFile(cache, data, 0644)
}

func load(v interface{}, file, key string, encrypted bool) error {
	cache := filepath.Join(Cache, file)

	_, err := os.Stat(cache)
	if err != nil {
		return err
	}

	// file exists
	data, err := os.ReadFile(cache)
	if err != nil {
		return err
	}

	if len(data) == 0 {
		return nil
	}

	if encrypted {
		val := decrypt(string(data), Key)
		data = []byte(val)
	}

	return json.Unmarshal(data, v)
}

// Backoff is for exponential backoff
func Backoff(attempts int) time.Duration {
	if attempts > 13 {
		return time.Hour
	}
	return time.Duration(math.Pow(float64(attempts), math.E)) * time.Millisecond * 100
}

// ID generates a new uuid
func ID() string {
	return uuid.New().String()
}

// Encrypt text using AES-256 and secret key
func Encrypt(text string) string {
	return encrypt(text, Key)
}

// Decrypt text using AES-256 and secret key
func Decrypt(text string) string {
	return decrypt(text, Key)
}

// Save data to the cache
func Save(data interface{}, file string, encrypt bool) error {
	return save(data, file, Key, encrypt)
}

// Load data from cache
func Load(data interface{}, file string, decrypt bool) error {
	return load(data, file, Key, decrypt)
}

// Render the template
func Render(w http.ResponseWriter, tmpl string) error {
	_, err := w.Write([]byte(tmpl))
	return err
}

// The standard HTML template
func Template(name, desc, nav, content string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
  <title>%s - %s | Micro</title>
  <meta name="description" content="%s">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <link rel="stylesheet" href="/mu.css">
  <style>
  body {
	  font-family: arial;
	  font-size: 14px;
	  margin: 0 auto;
	  max-width: 1600px;
  }
  a { color: black; text-decoration: none; }
  button:hover { cursor: pointer; }
  .anchor {
    top: -75px;
    margin-top: 75px;
    visibility: hidden;
    position: relative;
    display: block;

  }
  .category {
    font-weight: bold;
    font-size: small;
    padding: 5px;
    background: whitesmoke;
  }
  .headline {
    margin-bottom: 25px;
  }
  #info { margin-top: 5px;}
  #nav {
    position: fixed; background: white;
    padding: 25px;
    width: 20%%;
    margin-right: 50px;
    padding-top: 100px;
    vertical-align: top;
    display: inline-block;
    z-index: 100;
  }
  #content { display: block; height: 100%%; width: 70%%; margin-left: 30%%; display: inline-block; }
  #logo { margin-bottom: 25px; }
  .head { margin-right: 10px; font-weight: bold; }
  a.head { display: block; margin-bottom: 20px; }
  .section { display: block; max-width: 600px; margin-right: 20px; vertical-align: top;}
  .section img { display: none; }
  .section h3 { margin-bottom: 5px; }
  .ticker { display: inline-block; margin-right: 10px; }
  @media only screen and (max-width: 600px) {
    .section { margin-right: 0px; }
    #nav {
      position: fixed;
      padding: 20px;
      margin-right: 0;
      display: block;
      width: calc(100vw - 40px);
      overflow-x: scroll; white-space: nowrap;
      text-align: left;
    }
    #content {
      width: auto;
      padding: 20px;
      display: block;
      margin-left: 0;
    }
    a.head {
      display: inline-block;
      margin-bottom: 0;
    }
    #logo {
      margin-right: 10px;
      margin-bottom: 0;
      display: inline-block;
      vertical-align: middle;
    }
  }
  </style>
</head>
<body>
  <div id="nav">
    <div id="logo"><a href="/"><img height=40px width=auto src="/assets/mu.png"></a></div>
    %s
  </div>
  <div id="content">%s</div>
</body>
</html>
`, name, desc, desc, nav, content)

}

func Serve(port int) error {
	sub, _ := fs.Sub(html, "html")

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			// logged in
			sess, err := r.Cookie("sess")
			if err == nil && len(sess.Value) > 0 {
				http.Redirect(w, r, "/home", 302)
				return
			}
		}

		http.FileServer(http.FS(sub)).ServeHTTP(w, r)
	})

	if v := os.Getenv("PORT"); len(v) > 0 {
		port, _ = strconv.Atoi(v)
	}

	addr := fmt.Sprintf(":%d", port)

	return http.ListenAndServe(addr, nil)
}
