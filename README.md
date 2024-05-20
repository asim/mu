# mu

A Micro app platform

## Overview

Mu is an app platform that provides a simple set of building blocks for life.

The current list of apps:

- Chat - Channel based AI chat
- News - Topic based news feed
- Pray - Islamic prayer times
- Reminder - The Quran in English

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
