interface SubmissionCreateResponse {
  submission_id: string;
}

interface SubmissionStatusResponse {
  state: 'PENDING' | 'STARTED' | 'SUCCESS';
  status_msg: string;
  status_code: number;
  total_correct: number;
  total_testcases: number;
  display_runtime: string;
  runtime_percentile: number;
  memory: number;
  memory_percentile: number;
  lang: string;
  submission_id: string;
  task_finish_time: number;
}

function main() {
  console.log('[Code Duels] Network monitor injected into page world.');

  let problemSlug = '';
  const pathParts = window.location.pathname.split('/');
  if (pathParts[1] === 'problems' && pathParts[2]) {
    problemSlug = pathParts[2];
  }

  let currentSubmissionId = '';
  let awaitingSubmissionResponse = false;

  const originalFetch = window.fetch;

  window.fetch = async function(...args: [RequestInfo | URL, RequestInit | undefined]) {
    const resource = args[0];
    let rawUrl: string;
    if (resource instanceof Request) {
      rawUrl = resource.url;
    } else if (typeof resource === 'string') {
      rawUrl = resource;
    } else {
      rawUrl = resource.href; // URL object
    }
    const absoluteUrl = new URL(rawUrl, window.location.href).href;

    const submitUrlPattern = `https://leetcode.com/problems/${problemSlug}/submit/`;
    const checkUrlPattern = `https://leetcode.com/submissions/detail/${currentSubmissionId}/check/`;

    const response = await originalFetch.apply(this, args);

    if (absoluteUrl === submitUrlPattern) {
      try {
        const data: SubmissionCreateResponse = await response.clone().json();
        if (data.submission_id) {
          console.log('[Code Duels] Captured submission ID:', data.submission_id);
          currentSubmissionId = data.submission_id;
          awaitingSubmissionResponse = true;
        }
      } catch (e) {
        console.error('Error parsing submission response:', e);
      }
    }

    if (absoluteUrl === checkUrlPattern && awaitingSubmissionResponse) {
      try {
        const data: SubmissionStatusResponse = await response.clone().json();
        if (data.state === 'SUCCESS') {
          console.log('[Code Duels] Captured successful submission result:', data);
          awaitingSubmissionResponse = false;

          // Send the captured data back to the content script
          window.postMessage({
            source: 'leetcode-duel-network-monitor',
            type: 'submissionResult',
            data: {
              SubmissionID: parseInt(data.submission_id, 10),
              PassedTestCases: data.total_correct,
              TotalTestCases: data.total_testcases,
              Status: data.status_msg,
              Runtime: parseInt(data.display_runtime, 10) || null,
              RuntimePercentile: data.runtime_percentile,
              Memory: data.memory,
              MemoryPercentile: data.memory_percentile,
              Lang: data.lang,
              FinishTime: data.task_finish_time
                ? new Date(
                    // If task_finish_time is in seconds, multiply by 1000
                    data.task_finish_time > 1e12
                      ? data.task_finish_time
                      : data.task_finish_time * 1000
                  ).toISOString()
                : null,
            }
          }, '*');
        }
      } catch (e) {
        console.error('Error parsing submission status:', e);
      }
    }

    return response;
  };
}

(function() {
  if (window.__NETWORK_MONITOR_INJECTED__) {
    return;
  }
  window.__NETWORK_MONITOR_INJECTED__ = true;
  main();
})();
