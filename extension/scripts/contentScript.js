// Reference matchmake.js file for sendmatchmakinginfo function

// Content script to inject a div
(function() {
  // Create the new div element
  const newDiv = document.createElement('div');
  newDiv.innerHTML = "<strong>No stats yet</strong>";
  newDiv.style.position = 'fixed';
  newDiv.style.bottom = '20px';
  newDiv.style.right = '20px';
  newDiv.style.backgroundColor = 'lightblue';
  newDiv.style.padding = '10px';
  newDiv.style.borderRadius = '5px';
  newDiv.style.boxShadow = '0 2px 4px rgba(0,0,0,0.2)';
  newDiv.id = 'current-stats';

  // Append the new div to the body of the document
  document.body.appendChild(newDiv);
  console.log('Content script loaded');
  setTimeout(function() {
    const targetButtons = document.querySelectorAll('[data-e2e-locator="console-submit-button"]');
    console.log(targetButtons);
    // make the button console.log the text of the button on click
    targetButtons.forEach(function(button) {
      button.addEventListener('click', function() {
        console.log('Button clicked');
        // wait 5 seconds, then find spans with this class `class="text-sd-foreground text-lg font-semibold"`
        setTimeout(function() {
          const stats = document.getElementsByClassName('text-sd-foreground text-lg font-semibold');
          console.log(stats);
          const statsArray = Array.from(stats);
          const statsText = statsArray.map(stat => stat.innerText);
          const textWithPercentage = Array.from(document.querySelectorAll('span')).filter(span => span.innerText.includes('%'));
          console.log('.');
          console.log(textWithPercentage);
          // Get the percentage value from the two spans in the textWithPercentage array
          const percentValue1 = textWithPercentage[0].innerHTML;
          const percentValue2 = textWithPercentage[1].innerHTML;

          console.log(statsText);
          // Update the div with the new stats
          const currentStatsDiv = document.getElementById('current-stats');
          currentStatsDiv.innerHTML = `<strong>Runtime: ${statsText[0]}ms (${percentValue1}) | Memory: ${statsText[1]}MB (${percentValue2})</strong>`;

          // Call the function in matchmake.js
          sendMatchmakingInfo(statsText[0], statsText[1], percentValue1, percentValue2);
        }, 7000);
      });
    });
  }, 5000);

})();
function sendMatchmakingInfo(runtime, memory, percentValue1, percentValue2) {
  // Send the runtime, memory, and percentage values to the backend
  console.log("Sending stats to backend...");
  fetch('https://localhost:3000/matchmakeinfo', { // Endpoint not yet available
    method: 'POST',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      runtime: runtime,
      memory: memory,
      percentValue1: percentValue1,
      percentValue2: percentValue2
    })
  })
  .then(response => response.json())
  .then(data => {
    if (data.success) {
      console.log("Stats sent successfully");
    } else {
      console.log("Failed to send stats:", data.error);
    }
  })
  .catch(error => {
    console.error("Stats send error:", error);
  });
}