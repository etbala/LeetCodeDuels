// Content script to inject a div
(function() {
  // Create the new div element
  const newDiv = document.createElement('div');
  newDiv.innerHTML = "<strong>Opponent: 76% runtime | 33% Memory</strong>";
  newDiv.style.position = 'fixed';
  newDiv.style.bottom = '20px';
  newDiv.style.right = '20px';
  newDiv.style.backgroundColor = 'lightblue';
  newDiv.style.padding = '10px';
  newDiv.style.borderRadius = '5px';
  newDiv.style.boxShadow = '0 2px 4px rgba(0,0,0,0.2)';

  // Append the new div to the body of the document
  document.body.appendChild(newDiv);
  console.log('Content script loaded');
  
  const targetButtons = document.getElementById('ide-top-btns');
  // make the buttons blue
  targetButtons.style.display = 'none';
  for (let i = 0; i < 3; i++) {
    console.log('Button found');
    // targetButtons[i].addEventListener("click", function() {
    //   // This function will be executed when the button is clicked
    //   // You can add your desired actions here, like sending a message to the background script
    //   chrome.runtime.sendMessage({ message: "buttonClicked" });
    //   console.log('Button clicked');
    // });
  }
  
  
})();


