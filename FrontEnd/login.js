function hashText(inputText) {
    var hashedText = CryptoJS.SHA256(inputText);
    return hashedText;
  }

  function createEndPoint() {
    var password = document.getElementById("password").value;
    var username = document.getElementById("username").value;
    var endpoint = "/account/" + username + "/" + hashText(password + username);
    
    // Add query parameters to URL
    var url = "account.html?username=" + encodeURIComponent(username) + "&hash=" + encodeURIComponent(endpoint);
    
    // Open account.html page with query parameters
    window.open(url, "_self");
  }
  
  
