# mu

A Micro app platform

## Overview

Mu is a Micro app platform that provides a simple set of building blocks for life. It was born out of frustration with existing services using Ads or creating addictive behaviours. The goal is to create a separate system that addresses daily needs and nothing else. Part of that 
is focusing on Muslim related requirements like prayer times, the quran in english, etc. Other parts are related to dev interests like 
hacker news, open source, crypto.

## Apps

The current list of apps:

- **Chat** - General knowledge AI chat
- **News** - News headlines and market data
- **Pray** - Islamic prayer times
- **Reminder** - The holy Quran in English
- **Watch** - Watch YouTube videos
  
## Dependencies

- Go toolchain

## Usage

Download source

```bash
go install mu.dev/cmd/mu@latest
```

Run it

```
mu
```

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

## Development

Join [Discord](https://mu.xyz/discord) if interested.

## Live

Go to [https://mu.xyz](https://mu.xyz) to use the app.
