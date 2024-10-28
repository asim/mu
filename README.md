# mu

A muslim app platform

## Overview

Mu is an app platform that provides a simple set of building blocks for life. It was born out of a frustration with existing services. 
Most tech platforms create addictive behaviours through their algorithms. The intention is to create a separate system that fixes this.

## Apps

The current list of apps:

- **Chat** - General knowledge AI Q&A
- **News** - Hadiths, headlines and crypto
- **Pray** - Prayer times by location 
- **Reminder** - Read the Quran in English
- **Watch** - Search for YouTube videos
  
## Dependencies

- Go toolchain

## Usage

Download source

```bash
go install github.com/asim/mu/cmd/mu@latest
```

Run it

```
mu
```

Then go to `http://localhost:8080`

## Admin

A basic user admin on `/admin` displays the users. It requires `USER_ADMIN` to be set to the user who can view it

```
export USER_ADMIN=asim
```

Goto `localhost:8080`
## APIs

Set `OPENAI_API_KEY` from `openai.com` for ability to chat with AI

```
export OPENAI_API_KEY=xxx
```

Set `SUNNAH_API_KEY` from `sunnah.com` for daily hadith in news app

```
export SUNNAH_API_KEY=xxx
```

Set `CRYPTO_API_KEY` from `cryptocompare.com` for crypto market tickers

```
export CRYPTO_API_KEY=xxx
```

Set `YOUTUBE_API_KEY` from [Google Cloud](https://console.cloud.google.com/apis/api/youtube.googleapis.com/credentials) for YouTube data

```
export YOUTUBE_API_KEY
```

## PWA

Mu operates as a progressive web app. The main app can be installed just like a native app. 

Caching and offline mode is still a WIP.
