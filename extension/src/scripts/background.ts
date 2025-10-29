import { BackgroundAction, BackgroundActionType } from 'app/models/background-actions';
import { environment } from '../environments/environment';
import { ExtensionEventType, UIMessage, StartGamePayload } from 'app/models/extension-events';
import { ServerMessageType } from 'app/models/server-messages';
import { AUTH_TOKEN_KEY, USER_KEY } from 'app/common/auth.constants';

interface ServerMessage {
  type: string;
  payload: unknown;
}

let socket: WebSocket | null = null;
const API_BASE_URL = environment.apiUrl;
const SOCKET_URL = API_BASE_URL.replace(/^http/, 'ws');

const WEBSOCKET_KEEP_ALIVE_ALARM = 'websocket-keep-alive';
const WEBSOCKET_RECONNECT_ALARM = 'websocket-reconnect';

const RECONNECT_BASE_DELAY_MS = 1000;
const RECONNECT_MAX_DELAY_MS = 60000;
const RECONNECT_FACTOR = 2;

let isIntentionalDisconnect = false;

async function logout() {
  await chrome.storage.local.remove([AUTH_TOKEN_KEY, USER_KEY]);
}

// Clears any pending reconnect alarms and resets the attempt counter.
async function resetReconnectState() {
  await chrome.storage.local.remove('reconnectAttempt');
  chrome.alarms.clear(WEBSOCKET_RECONNECT_ALARM);
}

// Calculates delay and schedules next reconnect attempt
async function scheduleReconnect() {
  try {
    const { reconnectAttempt = 0 } = await chrome.storage.local.get('reconnectAttempt');

    // Calculate delay: base * (factor ^ attempt)
    let delayMs = RECONNECT_BASE_DELAY_MS * Math.pow(RECONNECT_FACTOR, reconnectAttempt);

    // Add jitter: +/- 500ms to prevent thundering herd
    const jitter = (Math.random() - 0.5) * 1000;
    delayMs = delayMs + jitter;

    delayMs = Math.min(delayMs, RECONNECT_MAX_DELAY_MS);
    delayMs = Math.max(1000, delayMs);

    console.log(`Scheduling reconnect attempt ${reconnectAttempt + 1} in ${Math.round(delayMs / 1000)}s`);
    
    await chrome.storage.local.set({ reconnectAttempt: reconnectAttempt + 1 });

    chrome.alarms.create(WEBSOCKET_RECONNECT_ALARM, {
      delayInMinutes: delayMs / 60000 
    });

  } catch (error) {
    console.error("Error scheduling reconnect:", error);
  }
}

async function connectWebSocket(): Promise<{ status: string; message?: string }> {
  if (socket && socket.readyState === WebSocket.OPEN) {
    console.log("WebSocket is already connected.");
    return { status: "success", message: "WebSocket is already connected." };
  }

  isIntentionalDisconnect = false;

  try {
    const storage = await chrome.storage.local.get(AUTH_TOKEN_KEY);
    const token = storage[AUTH_TOKEN_KEY];

    if (!token) {
      console.log("No auth token found. WebSocket will not connect.");
      return { status: "error", message: "No auth token found." };
    }

    const ticketResponse = await fetch(`${API_BASE_URL}/api/v1/ws-ticket`, {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${token}`
      }
    });

    if (ticketResponse.status === 401) {
      console.warn("Authentication token is invalid or expired. Logging out.");
      await logout();
      return { status: "error", message: "Authentication failed. Logged out." };
    }

    if (!ticketResponse.ok) {
      throw new Error(`Failed to get WebSocket ticket: ${ticketResponse.statusText}`);
    }

    const { ticket } = await ticketResponse.json();
    if (!ticket) {
        throw new Error("Ticket was not found in the server response.");
    }

    const socketUrlWithTicket = `${SOCKET_URL}/ws?ticket=${ticket}`;
    socket = new WebSocket(socketUrlWithTicket);

    socket.onopen = () => {
      console.log("WebSocket connection established securely using a ticket.");
      resetReconnectState();
      chrome.alarms.create(WEBSOCKET_KEEP_ALIVE_ALARM, {
        periodInMinutes: 0.45 // Fire every 27 seconds (prevent idle timeout)
      });
      setupSocketListeners();
    };

    socket.onerror = (err) => {
      console.error("WebSocket connection failed to open:", err);
    };

    return { status: "success", message: "WebSocket connection initiated." };

  } catch (error: unknown) {
    const errorMessage = error instanceof Error ? error.message : String(error);
    console.error("Failed to connect WebSocket:", errorMessage);
    return { status: "error", message: errorMessage };
  }
}

function disconnectWebSocket() {
    if (socket && socket.readyState === WebSocket.OPEN) {
        console.log("Disconnecting WebSocket.");
        isIntentionalDisconnect = true;
        chrome.alarms.clear(WEBSOCKET_KEEP_ALIVE_ALARM);
        chrome.alarms.clear(WEBSOCKET_RECONNECT_ALARM);
        socket.close();
        socket = null;
    }
}

chrome.alarms.onAlarm.addListener((alarm) => {
  if (alarm.name === WEBSOCKET_KEEP_ALIVE_ALARM) {
    sendToServer(ServerMessageType.ClientHeartbeat);
  } else if (alarm.name === WEBSOCKET_RECONNECT_ALARM) {
    (async () => {
      const result = await connectWebSocket();
      if (result.status === 'error') {
        scheduleReconnect();
      }
    })();
  }
});

function forwardToUI(type: ExtensionEventType, payload: unknown) {
  const message: UIMessage<unknown> = { type, payload };
  chrome.runtime.sendMessage(message);
}

function handleServerMessage(serverMsgData: ServerMessage) {
  console.log(`Received message from server: ${serverMsgData.type}`, serverMsgData.payload);

  const eventType = serverMsgData.type as ExtensionEventType;
  switch (eventType) {
    case ExtensionEventType.InvitationRequest:
      forwardToUI(eventType, serverMsgData.payload);
      break;

    case ExtensionEventType.StartGame: {
      const startPayload = serverMsgData.payload as StartGamePayload;
      console.log("Starting game, opening URL:", startPayload.problemURL);
      if (startPayload.problemURL) {
        chrome.tabs.create({ url: startPayload.problemURL });
      }
      forwardToUI(eventType, startPayload);
      break;
    }
    case ExtensionEventType.GameOver:
      forwardToUI(eventType, serverMsgData.payload);
      break;

    case ExtensionEventType.OpponentSubmission:
      forwardToUI(eventType, serverMsgData.payload);
      break;

    case ExtensionEventType.InvitationCanceled:
    case ExtensionEventType.InvitationDeclined:
    case ExtensionEventType.UserOffline:
      forwardToUI(eventType, serverMsgData.payload);
      break;
    
    case ExtensionEventType.InvitationNonexistant:
    case ExtensionEventType.OtherLogon:
      forwardToUI(eventType, null); // Events with no payload
      if (eventType === ExtensionEventType.OtherLogon) {
        disconnectWebSocket();
      }
      break;
    
    case ExtensionEventType.ServerError:
      console.error("Server-side error:", serverMsgData.payload);
      forwardToUI(eventType, serverMsgData.payload);
      break;

    default:
      console.warn(`Unhandled message type from server: ${serverMsgData.type}`);
      break;
  }
}

function setupSocketListeners() {
  if (!socket) return;

  socket.onerror = (err) => console.error("WebSocket error:", err);
  
  socket.onclose = (event) => {
    console.log(`WebSocket connection closed. Code: ${event.code}, Reason: ${event.reason}`);

    chrome.alarms.clear(WEBSOCKET_KEEP_ALIVE_ALARM);
    socket = null;

    if (isIntentionalDisconnect) {
      isIntentionalDisconnect = false;
      return;
    }

    console.log("Unexpected disconnect. Attempting to reconnect...");
    scheduleReconnect();
  };

  socket.onmessage = (event) => {
    try {
      const parsedData: unknown = JSON.parse(event.data);
      const message = parsedData as ServerMessage;
      if (!message) {
        throw new Error("Invalid message format");
      }
      handleServerMessage(message);
    } catch (e) {
      console.error("Could not parse server message:", event.data, e);
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

async function sendToServer(type: ServerMessageType, payload?: unknown) {
  if (socket && socket.readyState === WebSocket.OPEN) {
    const message = JSON.stringify({ type, payload });
    socket.send(message);
    return { status: "success", message: `Sent '${type}' to server.` };
  }
  
  if (type === ServerMessageType.ClientHeartbeat) {
    return { status: "error", error: "WebSocket is not connected. Skipping heartbeat." };
  }

  console.warn(`WebSocket not open. State: ${socket?.readyState}. Attempting reconnect for action: ${type}`);
  const storage = await chrome.storage.local.get(AUTH_TOKEN_KEY);
  const token = storage[AUTH_TOKEN_KEY];

  if (!token) {
    console.warn("Message not sent. User is unauthenticated.");
    return { status: "error", message: "No auth token found." };
  }

  await resetReconnectState();
  const connectResult = await connectWebSocket();

  if (connectResult.status === "success") {
    // wait a moment for connection to establish
    await new Promise((resolve) => setTimeout(resolve, 100));
    if (socket && socket.readyState === WebSocket.OPEN) {
      const message = JSON.stringify({ type, payload });
      socket.send(message);
      return { status: "success", message: `Sent '${type}' to server.` };
    }
  }

  return { status: "error", error: "WebSocket is not connected." };
}

// Listen for messages from Angular UI
chrome.runtime.onMessage.addListener((message: BackgroundAction, sender, sendResponse) => {
  console.log("Background received action:", message.action);
  switch (message.action) {
    case BackgroundActionType.DuelSendInvitation:
      sendResponse(sendToServer(ServerMessageType.ClientSendInvitation, message.payload));
      break;
    case BackgroundActionType.DuelAcceptInvitation:
      sendResponse(sendToServer(ServerMessageType.ClientAcceptInvitation, message.payload));
      break;
    case BackgroundActionType.DuelDeclineInvitation:
      sendResponse(sendToServer(ServerMessageType.ClientDeclineInvitation, message.payload));
      break;
    case BackgroundActionType.DuelCancelInvitation:
      sendResponse(sendToServer(ServerMessageType.ClientCancelInvitation));
      break;
    case BackgroundActionType.DuelSubmission:
      sendResponse(sendToServer(ServerMessageType.ClientSubmission, message.payload));
      break;
    case BackgroundActionType.DuelForfeit:
      sendResponse(sendToServer(ServerMessageType.ClientForfeit));
      break;
    default:
      sendResponse({ status: "error", error: "Unknown action" });
      break;
  }
  return true; // Keep the message channel open for async response
});

(function initialize() {
    connectWebSocket();
})();