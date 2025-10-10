import { BackgroundAction, BackgroundActionType } from 'app/models/background-actions';
import { environment } from '../environments/environment';
import { ExtensionEventType, UIMessage, StartGamePayload, InvitationRequestPayload, GameOverPayload, OpponentSubmissionPayload, UserIdentifiable } from 'app/models/extension-events';

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

    const ticketResponse = await fetch(`${API_BASE_URL}/api/v1/ws-ticket`, {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${token}`
      }
    });

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
      setupSocketListeners();
    };

    socket.onerror = (err) => {
      console.error("WebSocket connection failed to open:", err);
    };

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

function forwardToUI<T>(type: ExtensionEventType, payload: T) {
  console.log(`Forwarding event to UI: ${type}`, payload);
  const message: UIMessage<T> = {
    type: type,
    payload: payload,
  };
  chrome.runtime.sendMessage(message);
}

function handleServerMessage(serverMsgData: { type: string, payload: any }) {
  console.log(`Received message from server: ${serverMsgData.type}`, serverMsgData.payload);

  const eventType = serverMsgData.type as ExtensionEventType;
  switch (eventType) {
    case ExtensionEventType.InvitationRequest:
      forwardToUI<InvitationRequestPayload>(eventType, serverMsgData.payload);
      break;

    case ExtensionEventType.StartGame:
      const startPayload = serverMsgData.payload as StartGamePayload;
      console.log("Starting game, opening URL:", startPayload.problemURL);
      if (startPayload.problemURL) {
        chrome.tabs.create({ url: startPayload.problemURL });
      }
      forwardToUI<StartGamePayload>(eventType, startPayload);
      break;

    case ExtensionEventType.GameOver:
      forwardToUI<GameOverPayload>(eventType, serverMsgData.payload);
      break;

    case ExtensionEventType.OpponentSubmission:
      forwardToUI<OpponentSubmissionPayload>(eventType, serverMsgData.payload);
      break;

    case ExtensionEventType.InvitationCanceled:
    case ExtensionEventType.InvitationDeclined:
    case ExtensionEventType.UserOffline:
      forwardToUI<UserIdentifiable>(eventType, serverMsgData.payload);
      break;
    
    case ExtensionEventType.InvitationNonexistant:
    case ExtensionEventType.OtherLogon:
      forwardToUI<null>(eventType, null); // Events with no payload
      if (eventType === ExtensionEventType.OtherLogon) {
        disconnectWebSocket();
      }
      break;
    
    case ExtensionEventType.ServerError:
      console.error("Server-side error:", serverMsgData.payload);
      forwardToUI<any>(eventType, serverMsgData.payload);
      break;

    default:
      console.warn(`Unhandled message type from server: ${serverMsgData.type}`);
      break;
  }
}

function setupSocketListeners() {
  if (!socket) return;

  socket.onerror = (err) => console.error("WebSocket error:", err);
  
  socket.onclose = () => {
    console.log("WebSocket connection closed.");
    socket = null;
  };

  socket.onmessage = (event) => {
    try {
      const serverMsgData = JSON.parse(event.data);
      handleServerMessage(serverMsgData);
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

function sendToServer(type: BackgroundActionType, payload?: any) {
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
chrome.runtime.onMessage.addListener((message: BackgroundAction, sender, sendResponse) => {
  console.log("Background received action:", message.action);
  switch (message.action) {
    case BackgroundActionType.DuelSendInvitation:
      sendResponse(sendToServer(message.action, message.payload));
      break;
    case BackgroundActionType.DuelAcceptInvitation:
      sendResponse(sendToServer(message.action, message.payload));
      break;
    case BackgroundActionType.DuelDeclineInvitation:
      sendResponse(sendToServer(message.action, message.payload));
      break;
    case BackgroundActionType.DuelCancelInvitation:
      sendResponse(sendToServer(message.action));
      break;
    case BackgroundActionType.DuelSubmission:
      sendResponse(sendToServer(message.action, message.payload));
      break;
    default:
      sendResponse({ status: "error", error: "Unknown action" });
      break;
  }
  return true;
});

(function initialize() {
    connectWebSocket();
})();