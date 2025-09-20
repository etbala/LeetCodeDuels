# LeetCode Duels

![Server CI](https://github.com/etbala/LeetCodeDuels/actions/workflows/go.yml/badge.svg?branch=main)
![Extension CI](https://github.com/etbala/LeetCodeDuels/actions/workflows/extension.yml/badge.svg?branch=main)
![Deploy Server](https://github.com/etbala/LeetCodeDuels/actions/workflows/fly.yml/badge.svg?branch=main)

A browser extension that lets you challenge friends to real-time LeetCode duels.

## Getting Started

#### Clone the repository
```bash
git clone https://github.com/etbala/LeetCodeDuels.git
cd LeetCodeDuels
```

## Extension

### 1. Install Dependencies
Install Node.js (v22+) and Angular CLI
```bash
npm install -g @angular/cli
```

### 2. Build Extension
```bash
cd extension
npm install
ng build
```

Note: Build using `-c local` if you are trying to run the server locally:
```bash
ng build -c local
```

### 3. Load Extension in Browser
Load Unpacked Extension from Directory: `/extension/dist/browser`

> For Chrome, see https://support.google.com/chrome/a/answer/2714278

> For Firefox, see https://extensionworkshop.com/documentation/develop/temporary-installation-in-firefox/

## Server

### 1. Install Dependencies
Go (v1.24+)

### 2. Set up .env
Create a `.env` file based on the `.env.template`.
> Ensure file is located at `LeetCodeDuels/server/.env`

### 3. Run Server
```bash
cd server
go run ./cmd/server
```

## Testing

### Server Tests
```bash
cd server
go test ./... -v
```
