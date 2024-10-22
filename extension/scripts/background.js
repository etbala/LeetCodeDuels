

try {
    // ON page change
    chrome.tabs.onUpdated.addListener(function(tabId, changeInfo, tab) {
        if(changeInfo.status == 'complete'){
            chrome.scripting.executeScript({
                files: ['scripts/contentScript.js'],
                target: {tabId: tab.id}
            });
        }
    });

} catch(e) {
    console.log(e);
}

const SERVER_URL = 'http://localhost:8080';
chrome.runtime.onMessage.addListener((message, sender, sendResponse) => {
    if (message.type === 'getServerUrl') {
        sendResponse({ serverUrl: SERVER_URL });
    }
    return false;
});

chrome.webNavigation.onCompleted.addListener(function(details) {
    const url = details.url;

    // Check if this is your OAuth callback URL
    if (url.startsWith(`${SERVER_URL}/oauth/callback`)) {
        // Extract the authorization code from the URL
        const params = new URLSearchParams(new URL(url).search);
        const code = params.get('code');
        const state = params.get('state');

        if (code && state) {
            // Send the code and state to the backend for exchange
            fetch(`${SERVER_URL}/oauth/callback?code=${code}&state=${state}`)
            .then(response => {
                // Check for failure response
                if (!response.ok) {
                    return response.text().then(errorMessage => {
                        throw new Error(errorMessage);
                    });
                }
                return response.text();
            })
            .then(accessToken => {
                console.log("OAuth Token received:", accessToken);

                // Save access token
                chrome.storage.local.set({ "access_token": accessToken }, function() {
                    console.log("Access token saved");
                });

                // Close the tab
                chrome.tabs.remove(details.tabId);
            })
            .catch(error => {
                console.error('Error exchanging code:', error);
            });
        }
    }
}, {
    url: [{ urlContains: `${SERVER_URL}/oauth/callback` }]
});

