
/*
    Overrides the fetch() method to allow us to read data from api requests when
    we need to read it.

    Relevant APIs:
    https://leetcode.com/problems/{problem_slug}/submit/            -> Create submission, responds with submission id
    https://leetcode.com/submissions/detail/{submission_id}/check/  -> Checks the state/status of the submission
*/

(function() {
    let problemSlug = '';

    const pathParts = window.location.pathname.split('/');
    if (pathParts[1] === 'problems' && pathParts[2]) {
        problemSlug = pathParts[2];
    }

    let currentSubmissionId = '';
    let awaitingSubmissionResponse = false;
  
    // Override fetch
    const originalFetch = window.fetch;
    window.fetch = function(resource, init) {
      return originalFetch.apply(this, arguments).then(response => {
        const url = response.url;
  
        const submissionUrl = `https://leetcode.com/problems/${problemSlug}/submit/`;
        const statusUrl = `https://leetcode.com/submissions/detail/${currentSubmissionId}/check/`;
  
        // Submission creation
        if (url === submissionUrl) {
            response.clone().json().then(data => {
                console.log("Submission Response:", data)
                if (data.submission_id) {
                    currentSubmissionId = data.submission_id;
                    awaitingSubmissionResponse = true;
                }
            }).catch(error => {
                console.error('Error parsing submission response JSON:', error);
            });
        }
  
        // Submission status check
        else if (url === statusUrl && awaitingSubmissionResponse) {
            response.clone().json().then(data => {
                console.log("Submission State Response:", data)
                if (data.state === 'SUCCESS') {
                    
                    awaitingSubmissionResponse = false;
                    
                    const submissionData = {
                        PassedTestCases: data.total_correct,
                        TotalTestCases: data.total_testcases,
                        Status: data.status_msg,
                        Runtime: parseInt(data.status_runtime.replace(' ms', '')) || null,
                        Memory: data.memory || null,
                        Lang: data.lang,
                        SubmissionID: parseInt(data.submission_id) || null
                    };

                    // Send data to content script
                    window.postMessage({
                        source: 'leetcode-duel-network-monitor',
                        type: 'submissionResult',
                        data: submissionData
                    }, '*');

                }
          }).catch(error => {
              console.error('Error parsing submission status JSON:', error);
          });
        }
  
        return response;
      });
    };
})();
