# LeetCodeDuels

## Introduction
LeetCodeDuels is a project that allows users to engage in coding duels with LeetCode problems. This repository consists of a frontend extension, a backend server, and a scraper to interact with LeetCode.

## Repository Structure

### Frontend (Extension)
The frontend is built as a browser extension and includes HTML, CSS, and JavaScript files located in the `extension` directory:
- **HTML Files**: 
  - `hello.html` - Home page of the extension.
  - `login.html` - User login page.
  - `signup.html` - New user registration page.
- **CSS Files**:
  - `hello.css` - Styles for the home page.
  - `login.css` - Styles for the login page.
  - `signup.css` - Styles for the signup page.
- **JavaScript Files**:
  - `background.js` - Handles background tasks for the browser extension.
  - `contentScript.js` - Scripts that run in the context of web pages loaded by the browser.
  - `matchmake.js` - Handles the logic for matchmaking users for duels.

### Backend (Server)
The backend server is implemented in Go and handles API requests, located in the `server` directory:
- **API Handlers**:
  - `game/game.go` - Manages game logic and interactions.
  - `matchmaking/matchmaking.go` - Manages the matchmaking process.
- **Models**:
  - `pkg/models.go` - All of the models as structs
- **DB Queries**:
  - `pkg/` - Contains several files with different DB Query information
- **Main Entry**:
  - `cmd/server/main.go` - The main entry point for the backend server.

### Scraper
The scraper, located in the `scraper` directory, interacts with LeetCode to fetch problem data:
- `main.py` - Main script that runs the scraper.
- `utils.py` - Provides utility functions for the scraper.
- `writer.py` - Handles writing fetched data to the desired format.
