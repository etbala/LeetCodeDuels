# LeetCode Duels

![Server CI](https://github.com/etbala/LeetCodeDuels/actions/workflows/go.yml/badge.svg?branch=main)
![Extension CI](https://github.com/etbala/LeetCodeDuels/actions/workflows/extension.yml/badge.svg?branch=main)

A browser extension that lets you challenge friends to real-time LeetCode duels.

## Getting Started

### 1. Clone the repository
```bash
git clone https://github.com/etbala/LeetCodeDuels.git
cd LeetCodeDuels
```

### 2. Install Dependencies
Install Go (v1.24+) and Node.js (v22+).
```bash
npm install -g @angular/cli
```

### 3. Run Server
```bash
cd server/cmd/server
go run main.go
```

### 4. Compile Extension
```bash
cd extension
npm install
ng build
```

### 5. Load Extension in Browser
Load Unpacked Extension from Directory: `/extension/dist/browser`

> For Chrome, see https://support.google.com/chrome/a/answer/2714278

> For Firefox, see https://extensionworkshop.com/documentation/develop/temporary-installation-in-firefox/

## Testing

### Server Tests
```bash
cd server
go test ./... -v
```
