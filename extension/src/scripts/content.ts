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
  // This handles cases like "Accepted", "Wrong Answer", "Runtime Error", etc.
  const key = Object.keys(SubmissionStatus).find(k => SubmissionStatus[k as keyof typeof SubmissionStatus] === statusString);
  return key ? SubmissionStatus[key as keyof typeof SubmissionStatus] : SubmissionStatus.CompileError; // Default fallback
}

/**
 * Sends the formatted submission data to the background script for WebSocket communication.
 * @param submission - The PlayerSubmission object to send.
 */
function sendSubmissionToBackground(submission: PlayerSubmission): Promise<void> {
  return new Promise((resolve, reject) => {
    const payload: SubmissionPayload = { submission };
    const message: BackgroundAction<SubmissionPayload> = {
      action: BackgroundActionType.DuelSubmission,
      payload,
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
    playerID: 0, // Backend will assign this based on the authenticated user
    problemID: rawData.ProblemID,
    status: mapStatusToEnum(rawData.Status)
  };

  try {
    await sendSubmissionToBackground(playerSubmission);
  } catch (error) {
    console.error('[Code Duels] Failed to send submission to background script:', error);
  }
}

if ((window as any).__CODE_DUELS_EXTENSION_INJECTED__) {
} else {
  (window as any).__CODE_DUELS_EXTENSION_INJECTED__ = true;
  main();
}