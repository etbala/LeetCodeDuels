
// /*
// Relevant API Requests (that we want to read the data of):
//     https://leetcode.com/problems/{problem_slug}/submit/ (Create submission, gets submission id)
//     https://leetcode.com/submissions/detail/{submission_id}/check/ (Checks the state/status of the submission)
// */

(function() {
    // Variables
    let problemSlug = 'two-sum';
    let currentSubmissionId = '';
    let awaitingSubmissionResponse = false;
  
    // Listen for messages from the content script
    window.addEventListener('message', function(event) {
        if (event.source !== window || event.data.source !== 'leetcode-duel-extension') return;
    
        const message = event.data;
    
        if (message.type === 'updateVariables') {
            updateVariables(message.variables);
        }
    });
  
    // Helper function to send messages to content script
    function sendMessageToContentScript(message) {
      window.postMessage(Object.assign({ source: 'leetcode-duel-injected' }, message), '*');
    }
  
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
                // Update submission ID and notify content script
                currentSubmissionId = data.submission_id;
                awaitingSubmissionResponse = true;
    
                sendMessageToContentScript({
                    type: 'submissionCreated',
                    submissionId: currentSubmissionId
                });
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
                    // Submission result obtained
                    awaitingSubmissionResponse = false;
                    sendMessageToContentScript({
                        type: 'submissionResult',
                        data: data
                    });
                }
          }).catch(error => {
              console.error('Error parsing submission status JSON:', error);
          });
        }
  
        return response;
      });
    };

    // const originalXHROpen = XMLHttpRequest.prototype.open;
    // XMLHttpRequest.prototype.open = function(method, url, ...args) {
    //     this.addEventListener('readystatechange', function() {
    //         if (this.readyState === 4) { // DONE

    //             const submissionUrl = `https://leetcode.com/problems/${problemSlug}/submit/`;
    //             const statusUrl = `https://leetcode.com/submissions/detail/${currentSubmissionId}/check/`;

    //             // Check if the request is the submission POST request
    //             if (url === submissionUrl && method === 'POST') {
    //                 try {
    //                     const responseData = JSON.parse(this.responseText);
    //                     if (responseData.submission_id) {
    //                         submissionId = responseData.submission_id.toString();
    //                         awaitingSubmissionResponse = true;
    //                         console.log('Captured submission ID (XHR):', submissionId);
    //                     }
    //                 } catch (error) {
    //                     console.error('Error parsing submission response JSON (XHR):', error);
    //                 }
    //             }

    //             // Check if the request is the check submission status GET request
    //             if (awaitingSubmissionResponse && url === statusUrl) {
    //                 try {
    //                     const responseData = JSON.parse(this.responseText);
    //                     if (responseData.state === 'SUCCESS') {
    //                         awaitingSubmissionResponse = false;
    //                         // Send the data to the content script via window.postMessage
    //                         window.postMessage({ type: 'leetcodeSubmissionResult', data: responseData }, '*');
    //                     }
    //                 } catch (error) {
    //                     console.error('Error parsing check response JSON (XHR):', error);
    //                 }
    //             }
    //         }
    //     });

    //     originalXHROpen.apply(this, [method, url, ...args]);
    // };
  })();
