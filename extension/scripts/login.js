// document.getElementById('submitLogin').addEventListener('click', function(event) {
//     var username = document.getElementById('username').value;
//     var password = document.getElementById('password').value;

//     fetch('/login', {
//         method: 'POST',
//         headers: {
//             'Content-Type': 'application/json',
//         },
//         body: JSON.stringify({ username: username, password: password }),
//     })
//     .then(response => {
//         if (response.ok) {
//             // Check if the 'loggedIn' cookie is set
//             var cookies = document.cookie.split(';');
//             for (var i = 0; i < cookies.length; i++) {
//                 var cookie = cookies[i].trim();
//                 if (cookie.startsWith('loggedIn=true')) {
//                     // Redirect to another page or update UI as needed
//                     console.log('User logged in');
//                     return;
//                 }
//             }
//             // Handle case where cookie is not set
//             console.log('User not logged in');
//         } else {
//             throw new Error('Failed to login');
//         }
//     })
//     .catch(error => {
//         console.error('Error during login:', error);
//         // Handle login errors, show message to the user
//     });
// });

document.getElementById('loginButton').addEventListener('click', function () {
    chrome.tabs.create({
        url: 'http://localhost:8080/oauth/authorize'
    });
});