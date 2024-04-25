# Getting Started Guide for LeetCodeDuels

This guide will walk you through setting up the LeetCodeDuels project on your local machine. Follow the steps below to ensure you have everything you need to start contributing to the project.

## Prerequisites

Before you start, ensure you have the following software installed on your computer:

1. **Go**: Required for backend development.
   - Download and install Go from [golang.org](https://golang.org/dl/).

2. **Python**: Necessary for running the scraper.
   - Download and install Python from [python.org](https://www.python.org/downloads/).
   - Ensure `pip` is installed, as it is needed to install Python packages.

3. **Git**: Essential for version control and collaborating on the project.
   - Download and install Git from [git-scm.com](https://git-scm.com/downloads).

## Setup Instructions

### Cloning the Repository

Start by cloning the repository to your local machine:

```bash
git clone https://github.com/etbala/LeetCodeDuels.git
cd LeetCodeDuels
```

### Setting up the Backend

Navigate to the server directory and install Go dependencies:

```bash
cd server
go mod tidy
```

To start the server run
```bash
go run cmd/server/main.go
```

### Adding the Browser Extension

To add the extension to your browser for development:

Open your browser and navigate to the extensions management page:
1. Chrome: Enter chrome://extensions/ in the address bar. Firefox: Enter about:addons in the address bar.
2. Enable Developer Mode (usually a toggle in the top-right corner).
3. Click on "Load Unpacked" (Chrome) or "Load Temporary Add-on" (Firefox) and select the extension directory from the LeetCodeDuels repository.

## Using the extension

While most of the endpoints the extension needs to work are not yet implemented, there is functionality that currently comes out of it.

When you open a leetcode page, you will see the injected html coming from the extension saying "No Stats Yet".
<!-- Make a bullet point list -->
- Go ahead and open up a leetcode question of your choice.
- Complete the question as you normally would.
- Once you submit the question, the extension will update the html to show the stats of the question you just completed.

<!-- make this bold -->
While the extension is not fully functional, this is a small example of what it can do, and how we can use it to deliver information to the user from leeetcode.

### Setting up the Scraper
Navigate back to the scraper directory and install required Python packages:

```bash
cd scraper
pip install -r requirements.txt
```

To run the scraper, execute:
```bash
python main.py
```
