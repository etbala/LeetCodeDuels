import { BackgroundScriptMessage, ExtensionEvent } from './../app/models/background.models';
import { environment } from '../environments/environment';

let socket: WebSocket | null = null;
const API_BASE_URL = environment.apiUrl;
const AUTH_TOKEN_KEY = 'auth_token';
const SOCKET_URL = API_BASE_URL.replace(/^http/, 'ws');

async function connectWebSocket(): Promise<{ status: string; message?: string }> {
  if (socket && socket.readyState === WebSocket.OPEN) {
    console.log("WebSocket is already connected.");
    return { status: "success", message: "WebSocket is already connected." };
  }

  try {
    const storage = await chrome.storage.local.get(AUTH_TOKEN_KEY);
    const token = storage[AUTH_TOKEN_KEY];

    if (!token) {
      throw new Error("User is not authenticated. Cannot connect WebSocket.");
    }

    const response = await fetch(`${API_BASE_URL}/ws`, {
      headers: {
        'Authorization': `Bearer ${token}`
      }
    });

    // If this check fails, the server doesn't allow an upgrade.
    if (!response.ok) {
      throw new Error(`Authorization check failed with status: ${response.status}`);
    }

    socket = new WebSocket(SOCKET_URL);
    setupSocketListeners();

    return { status: "success", message: "WebSocket connection initiated." };

  } catch (error: any) {
    console.error("Failed to connect WebSocket:", error);
    return { status: "error", message: error.message };
  }
}

function disconnectWebSocket() {
    if (socket && socket.readyState === WebSocket.OPEN) {
        console.log("Disconnecting WebSocket.");
        socket.close();
        socket = null;
    }
}

function setupSocketListeners() {
  if (!socket) return;

  socket.onopen = () => console.log("WebSocket connection established.");
  socket.onerror = (err) => console.error("WebSocket error:", err);
  socket.onclose = () => {
    console.log("WebSocket connection closed.");
    socket = null; // Clear the socket instance on close
  };

  socket.onmessage = (event) => {
    try {
      const serverMsgData = JSON.parse(event.data);
      const eventToUi: ExtensionEvent = {
        event: serverMsgData.type,
        data: serverMsgData.payload,
      };
      // Forward the message from your server to the Angular UI
      chrome.runtime.sendMessage(eventToUi);
    } catch (e) {
      console.error("Could not parse server message:", event.data);
    }
  };
}

// Listen for changes to auth token
chrome.storage.onChanged.addListener((changes, areaName) => {
    if (areaName === 'local' && changes[AUTH_TOKEN_KEY]) {
        const storageChange = changes[AUTH_TOKEN_KEY];
        if (storageChange.newValue) {
            console.log("Auth token found. Attempting to connect WebSocket.");
            connectWebSocket();
        } else {
            console.log("Auth token removed. Disconnecting WebSocket.");
            disconnectWebSocket();
        }
    }
});

function sendToServer(type: string, payload?: any) {
  if (socket && socket.readyState === WebSocket.OPEN) {
    const message = JSON.stringify({ type, ...payload });
    socket.send(message);
    return { status: "success", message: `Sent '${type}' to server.` };
  } else {
    // todo: try reconnecting?
    return { status: "error", error: "WebSocket is not connected." };
  }
}

// Listen for messages from Angular UI
chrome.runtime.onMessage.addListener((message: BackgroundScriptMessage, sender, sendResponse) => {
  console.log("Background received action:", message.action);
  switch (message.action) {
    case 'duel:send-invitation':
      sendResponse(sendToServer("send_invitation", message.payload));
      break;
    case 'duel:accept-invitation':
      sendResponse(sendToServer("accept_invitation", message.payload));
      break;
    case 'duel:decline-invitation':
      sendResponse(sendToServer("decline_invitation", message.payload));
      break;
    case 'duel:cancel-invitation':
      sendResponse(sendToServer("cancel_invitation"));
      break;
    case 'duel:submission':
      sendResponse(sendToServer("submission", message.payload));
      break;
    default:
      sendResponse({ status: "error", error: "Unknown action" });
      break;
  }
  return true;
});

// Attempt to connect on initial script startup if already authenticated.
(function initialize() {
    console.log("Background script started. Checking auth status...");
    connectWebSocket();
})();