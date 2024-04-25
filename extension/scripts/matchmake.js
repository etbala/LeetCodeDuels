javascript
function handleClick() {
  // Send matchmaking request to backend
  fetch('https://your-backend-url/matchmake', { // Replace 
    method: 'POST',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      userId: getUserId() // Get the user's ID from storage or cookies
    })
  })
  .then(response => response.json())
  .then(data => {
    if (data.success) {
      // Match found! Get the question URL and redirect
      window.location.href = data.questionUrl;
    } else {
      // Handle case where no match is found or an error occurs
      console.log("Matchmaking failed:", data.error);
      // You can display an appropriate message to the user here
    }
  })
  .catch(error => {
    console.error("Matchmaking request error:", error);
    // Handle network errors or other unexpected issues
  });
}

function getUserId() {
  return fetch('https://your-backend-url/user', { // User API endpoint not yet available
    method: 'GET',
    headers: {
      'Content-Type': 'application/json'
    }
  })
  .then(response => response.json())
  .then(data => {
    return data.userId;
  })
  .catch(error => {
    console.error("Failed to get user ID:", error);
    // Handle network errors or other unexpected issues
  });
}

// onclick event listener for the each button from the extension popup
document.getElementById('easy').addEventListener('click', function() {
  console.log('Easy button clicked');
  findQuestion('easy');
});

document.getElementById('medium').addEventListener('click', function() {
  findQuestion('medium');
});

document.getElementById('hard').addEventListener('click', function() {
  findQuestion('hard');
});


function findQuestion(difficulty) {
  // Get the json from ../../scraper/lc_problems.json
  fetch('../../scraper/lc_problems.json')
  .then(response => response.json())
  .then(data => {
    // Filter the data based on the difficulty
    const questions = data.filter(question => question.difficulty === difficulty);
    console.log(questions);
    // Get a random question from the filtered data
    const randomQuestion = questions[Math.floor(Math.random() * questions.length)];
    // Redirect to the question URL
    window.location.href = randomQuestion.url;
  })
}