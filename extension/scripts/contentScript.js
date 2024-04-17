// Content script to inject a div
(function() {
  // Create the new div element
  const newDiv = document.createElement('div');
  newDiv.innerHTML = "<strong>Hello, this is the injected div!</strong>";
  newDiv.style.position = 'fixed';
  newDiv.style.bottom = '20px';
  newDiv.style.right = '20px';
  newDiv.style.backgroundColor = 'lightblue';
  newDiv.style.padding = '10px';
  newDiv.style.borderRadius = '5px';
  newDiv.style.boxShadow = '0 2px 4px rgba(0,0,0,0.2)';

  // Append the new div to the body of the document
  document.body.appendChild(newDiv);
})();