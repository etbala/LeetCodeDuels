
window.onload = function() {
  // Check if the user is logged in by looking for the JWT
  chrome.storage.local.get(['token', 'github_username'], function(data) {
      if (data.token && data.github_username) {
          // User is logged in, hide the login button and show user info
          document.getElementById('login-btn').style.display = 'none';
          document.getElementById('user-info').style.display = 'block';
          document.getElementById('username').textContent = data.github_username;
      } else {
          // User is not logged in, show the login button
          document.getElementById('login-btn').style.display = 'block';
          document.getElementById('user-info').style.display = 'none';
      }
  });
}

document.getElementById('login-btn').addEventListener('click', function () {
  const redirectUri = chrome.identity.getRedirectURL();
  const clientId = 'Ov23liQ4ERGhUYdeT8yb';
  const authURL = 'https://github.com/login/oauth/authorize' +
    '?client_id=' + encodeURIComponent(clientId) +
    '&redirect_uri=' + encodeURIComponent(redirectUri) +
    '&scope=user';

  chrome.identity.launchWebAuthFlow({
    url: authURL,
    interactive: true
  }, function (redirectUrl) {
    if (chrome.runtime.lastError) {
      console.error('Auth flow error:', chrome.runtime.lastError);
      return;
    }

    // Extract the authorization code from the redirect URL
    const params = new URLSearchParams(new URL(redirectUrl).search);
    const code = params.get('code');

    if (code) {
      // Send the code to your backend for exchange
      fetch('http://localhost:8080/oauth/exchange-token', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ code })
      })
      .then(response => {
        if (!response.ok) {
          return response.text().then(errorMessage => {
            throw new Error(errorMessage);
          });
        }
        return response.json();
      })
      .then(data => {
        const token = data.token; // The JWT received from the backend
        console.log("JWT received:", token);

        // Save the JWT
        chrome.storage.local.set({ "token": token }, function() {
          console.log("JWT saved");

          // Decode the JWT to get the user info (optional)
          const payloadBase64 = token.split('.')[1];
          const decodedPayload = JSON.parse(atob(payloadBase64));
          const githubUsername = decodedPayload.username;

          // Save the GitHub username in chrome storage
          chrome.storage.local.set({ "github_username": githubUsername }, function() {
            console.log("GitHub username saved:", githubUsername);

            // Update the UI
            document.getElementById('login-btn').style.display = 'none';
            document.getElementById('user-info').style.display = 'block';
            document.getElementById('username').textContent = githubUsername;
          });
        });
      })
      .catch(error => {
        console.error('Error exchanging code:', error);
      });
    } else {
      console.error('Authorization code not found in redirect URL');
    }
  });
});

function handleClick() {
  // Send matchmaking request to backend
  // fetch('https://your-backend-url/matchmake', { // Replace 
  //   method: 'POST',
  //   headers: {
  //     'Content-Type': 'application/json'
  //   },
  //   body: JSON.stringify({
  //     userId: getUserId() // Get the user's ID from storage or cookies
  //   })
  // })
  // .then(response => response.json())
  // .then(data => {
  //   if (data.success) {
  //     // Match found! Get the question URL and redirect
  //     window.location.href = data.questionUrl;
  //   } else {
  //     // Handle case where no match is found or an error occurs
  //     console.log("Matchmaking failed:", data.error);
  //     // You can display an appropriate message to the user here
  //   }
  // })
  // .catch(error => {
  //   console.error("Matchmaking request error:", error);
  //   // Handle network errors or other unexpected issues
  // });
}

// function getUserId() {
//   return fetch('https://your-backend-url/user', { // User API endpoint not yet available
//     method: 'GET',
//     headers: {
//       'Content-Type': 'application/json'
//     }
//   })
//   .then(response => response.json())
//   .then(data => {
//     return data.userId;
//   })
//   .catch(error => {
//     console.error("Failed to get user ID:", error);
//     // Handle network errors or other unexpected issues
//   });
// }

// function findQuestion(difficulty) {
//   // Get the json from ../../scraper/lc_problems.json
//   fetch('./lc_problems.json')
//   .then(response => response.json())
//   .then(data => {
//       // Filter the data based on the difficulty
//       console.log(data);
//       const questions = [];
//       for (let i = 0; i < 2278; i++) {
//         if (data[i].difficulty === difficulty) {
//           questions.push(data[i]);
//         }
//       }
//       console.log(questions);
//       // Get a random question from the filtered data
//       const randomQuestion = questions[Math.floor(Math.random() * questions.length)];
//       // Redirect to the question URL
//       window.location.href = randomQuestion.url;
//       console.log(randomQuestion.url);
//   })
//   .catch(error => {
//     console.error("Failed to fetch data:", error);
//     // Handle network errors or other unexpected issues
//   });
// }
//   // onclick event listener for the each button from the extension popup
//   console.log('Event listeners added');
//   document.getElementById('easy').addEventListener('click', function() {
//     console.log('Easy button clicked');
//     findQuestion('Easy');
//   });

//   document.getElementById('medium').addEventListener('click', function() {
//     findQuestion('Medium');
//   });

//   document.getElementById('hard').addEventListener('click', function() {
//     findQuestion('Hard');
//   });
  


