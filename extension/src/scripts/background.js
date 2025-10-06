import { environment } from '../environments/environment.ts';

try {
    // ON page change
    chrome.tabs.onUpdated.addListener(function(tabId, changeInfo, tab) {
        if(changeInfo.status == 'complete'){
            chrome.scripting.executeScript({
                files: ['scripts/content-script.js'],
                target: {tabId: tab.id}
            });
        }
    });

} catch(e) {
    console.log(e);
}

const SERVER_URL = environment.apiUrl;
chrome.runtime.onMessage.addListener((message, sender, sendResponse) => {
    if (message.type === 'getServerUrl') {
        sendResponse({ serverUrl: SERVER_URL });
    }
    return false;
});
