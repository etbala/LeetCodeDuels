document.getElementById("submitLogin").addEventListener("click", function(event) {
    event.preventDefault();
    var email = document.getElementById("email").value;
    var username = document.getElementById("username").value;
    var password = document.getElementById("password").value;
    var confirmPassword = document.getElementById("confirmPassword").value;
    
    // Validate password match
    if (password !== confirmPassword) {
        alert("Passwords do not match");
        return;
    }

    // Create user object
    var user = {
        email: email,
        username: username,
        password: password
    };

    // Send POST request to backend
    fetch("/createUser", {
        method: "POST",
        headers: {
            "Content-Type": "application/json"
        },
        body: JSON.stringify(user)
    })
    .then(response => {
        if (!response.ok) {
            throw new Error("Network response was not ok");
        }
        // Set cookie upon successful sign-up
        document.cookie = `loggedIn=true; path=/`; // Set a simple flag cookie
        return response.json();
    })
    .then(data => {
        console.log("User created successfully:", data);
        // Optionally, redirect to another page or show a success message
    })
    .catch(error => {
        console.error("Error creating user:", error);
    });
});
