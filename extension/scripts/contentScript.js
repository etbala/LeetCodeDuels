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
          console.log(statsText);
          // Update the div with the new stats
          const currentStatsDiv = document.getElementById('current-stats');
          currentStatsDiv.innerHTML = statsText.join('<br>');
        }, 7000);
      });
    });
  }, 5000);

})();


