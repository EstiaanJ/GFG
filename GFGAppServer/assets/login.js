function hashText(inputText) {
    var hashedText = CryptoJS.SHA256(inputText);
    return hashedText;
  }

  function createEndPoint() {
    var password = document.getElementById("password").value;
    var username = document.getElementById("username").value;
    
    // Validate user input
    if (!/^[a-zA-Z0-9]+$/.test(username)) {
        alert("Invalid username");
        return;
    }
    if (!/^[a-zA-Z0-9@#$%^&+=]+$/.test(password)) {
        alert("Invalid password");
        return;
    }
    
    // Encode user input
    var encodedUsername = encodeURIComponent(username);
    var encodedPassword = encodeURIComponent(password + username);
    
    // Hash password
    var hashedPassword = hashText(encodedPassword);
    
    // Build endpoint URL
    var endpoint = "/account/" + encodedUsername + "/" + hashedPassword;
    
    // Build query string
    var urlParams = new URLSearchParams();
    urlParams.append("username", encodedUsername);
    urlParams.append("hash", endpoint);
    
    // Build URL with query string
    var url = "account.html?" + urlParams.toString();
    
    // Open account.html page with query parameters
    window.open(url, "_self");
}
  
