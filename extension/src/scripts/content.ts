import { BackgroundAction, BackgroundActionType, SubmissionPayload } from '../app/models/background-actions';
import { LanguageType, PlayerSubmission, SubmissionStatus } from '../app/models/match';

interface InjectedScriptMessage {
  source: 'leetcode-duel-network-monitor';
  type: 'submissionResult';
  data: LeetCodeSubmissionData;
}

interface LeetCodeSubmissionData {
  SubmissionID: number;
  ProblemID: number,
  PassedTestCases: number;
  TotalTestCases: number;
  Status: string;
  Runtime: number;
  RuntimePercentile: number;
  Memory: number;
  MemoryPercentile: number;
  Lang: string;
  FinishTime: number;
}

let triggerButton: HTMLButtonElement | null = null;
let appFrame: HTMLIFrameElement | null = null;

function main() {
  console.log("[Code Duels] Content script injected.");

  const script = document.createElement('script');
  script.src = chrome.runtime.getURL('networkMonitor.js');
  (document.head || document.documentElement).appendChild(script);
  script.onload = () => script.remove();

  window.addEventListener('message', (event: MessageEvent<InjectedScriptMessage>) => {
    if (
      event.source === window &&
      event.data?.source === 'leetcode-duel-network-monitor' &&
      event.data?.type === 'submissionResult'
    ) {
      handleSubmissionResult(event.data.data);
    }
  });

  appFrame = document.createElement('iframe');
  appFrame.id = 'leetcode-duels-iframe';
  appFrame.src = chrome.runtime.getURL('index.html');
  document.body.appendChild(appFrame);

  triggerButton = document.createElement('button');
  triggerButton.id = 'leetcode-duels-trigger';
  triggerButton.innerText = 'D'; // Placeholder, can use icon
  document.body.appendChild(triggerButton);

  triggerButton.addEventListener('click', toggleAppFrame);

  chrome.runtime.onMessage.addListener(
    (message, sender, sendResponse) => {
      if (message.action === "toggle_ui") {
        toggleAppFrame();
        sendResponse({ status: "success" });
      }
      return true;
    }
  );

  window.addEventListener('message', (event) => {
    if (event.source === appFrame?.contentWindow && event.data?.action === 'toggleUI') {
      toggleAppFrame();
    }
  });
}

function toggleAppFrame() {
  if (!appFrame || !triggerButton) {
    console.error("App frame or trigger button does not exist.");
    return;
  }
  
  const isNowVisible = appFrame.classList.toggle('visible');
  triggerButton.style.display = isNowVisible ? 'none' : 'block';
}

/**
 * Maps the captured LeetCode language string to our LanguageType enum.
 * @param langString The language string from LeetCode (e.g., "python3").
 * @returns The corresponding LanguageType enum member.
 */
function mapLangToEnum(langString: string): LanguageType {
  const langKey = langString.charAt(0).toUpperCase() + langString.slice(1);
  return LanguageType[langKey as keyof typeof LanguageType] || LanguageType.Erlang; // Default fallback
}

/**
* Maps the captured LeetCode status string to our SubmissionStatus enum.
* @param statusString The status string from LeetCode (e.g., "Accepted").
* @returns The corresponding SubmissionStatus enum member.
*/
function mapStatusToEnum(statusString: string): SubmissionStatus {
  const key = Object.keys(SubmissionStatus).find(k => SubmissionStatus[k as keyof typeof SubmissionStatus] === statusString);
  return key ? SubmissionStatus[key as keyof typeof SubmissionStatus] : SubmissionStatus.OutputLimitExceeded; // Default fallback
}

/**
 * Sends the formatted submission data to the background script for WebSocket communication.
 * @param submission - The PlayerSubmission object to send.
 */
function sendSubmissionToBackground(submission: PlayerSubmission): Promise<void> {
  return new Promise((resolve, reject) => {
    const message: BackgroundAction<SubmissionPayload> = {
      action: BackgroundActionType.DuelSubmission,
      payload: submission,
    };

    chrome.runtime.sendMessage(message, (response) => {
      if (chrome.runtime.lastError) {
        return reject(chrome.runtime.lastError.message);
      }
      if (response?.status === 'error') {
        return reject(response.error);
      }
      console.log('[Code Duels] Submission successfully sent to background script.');
      resolve();
    });
  });
}

/**
 * Handles the 'submissionResult' message from the injected script.
 * It transforms the raw data into our PlayerSubmission model and sends it off.
 * @param rawData - The data captured by the network monitor.
 */
async function handleSubmissionResult(rawData: LeetCodeSubmissionData): Promise<void> {
  const playerSubmission: PlayerSubmission = {
    submissionID: rawData.SubmissionID,
    problemID: rawData.ProblemID,
    time: rawData.FinishTime ? new Date(rawData.FinishTime).toISOString() : new Date().toISOString(),
    status: mapStatusToEnum(rawData.Status),
    language: mapLangToEnum(rawData.Lang),
    passedTestCases: rawData.PassedTestCases,
    totalTestCases: rawData.TotalTestCases,
    runtime: rawData.Runtime,
    runtimePercentile: rawData.RuntimePercentile,
    memory: rawData.Memory,
    memoryPercentile: rawData.MemoryPercentile,
  };

  try {
    await sendSubmissionToBackground(playerSubmission);
  } catch (error) {
    console.error('[Code Duels] Failed to send submission to background script:', error);
  }
}

if (!window.__CODE_DUELS_EXTENSION_INJECTED__) {
  window.__CODE_DUELS_EXTENSION_INJECTED__ = true;
  main();
}