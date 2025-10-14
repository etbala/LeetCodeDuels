interface SubmissionCreateResponse {
  submission_id: string;
}

interface SubmissionStatusResponse {
  state: 'PENDING' | 'STARTED' | 'SUCCESS';
  status_msg: string;
  total_correct: number;
  total_testcases: number;
  status_runtime: string;
  memory: string;
  lang: string;
  submission_id: string;
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
    const resourceUrl = typeof resource === 'string' ? resource : (resource as any).url;

    const submitUrlPattern = `https://leetcode.com/problems/${problemSlug}/submit/`;
    const checkUrlPattern = `https://leetcode.com/submissions/detail/${currentSubmissionId}/check/`;

    const response = await originalFetch.apply(this, args);

    if (resourceUrl === submitUrlPattern) {
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

    if (resourceUrl === checkUrlPattern && awaitingSubmissionResponse) {
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
              SubmissionID: parseInt(data.submission_id, 10) || null,
              PassedTestCases: data.total_correct,
              TotalTestCases: data.total_testcases,
              Status: data.status_msg,
              Runtime: parseInt(data.status_runtime.replace(' ms', ''), 10) || null,
              Memory: parseFloat(data.memory.replace(' MB', '')) || null,
              Lang: data.lang,
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
  if ((window as any).__NETWORK_MONITOR_INJECTED__) {
    return;
  }
  (window as any).__NETWORK_MONITOR_INJECTED__ = true;
  main();
})();
