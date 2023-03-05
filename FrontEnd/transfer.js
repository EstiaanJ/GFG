

function transfer() {
  var to_acc_username = document.getElementById("to_acc_username").value;
  var amount = document.getElementById("amount").value;
  var from_acc_name = document.getElementById("username").textContent;

  // Validate user input
  if (!/^[a-zA-Z0-9]+$/.test(to_acc_username)) {
      alert("Invalid recipient account username");
      return;
  }
  if (isNaN(amount) || amount <= 0) {
      alert("Invalid transfer amount");
      return;
  }
  
  // Encode user input
  var encodedToUsername = encodeURIComponent(to_acc_username);
  var encodedAmount = encodeURIComponent(amount);
  var encodedFromUsername = encodeURIComponent(from_acc_name);
  
  // Make a POST request to the /transfer endpoint with the encoded user input
  $.ajax({
    url: "http://51.161.163.66:44658/transfer",
    type: "POST",
    dataType: "json",
    data: JSON.stringify({from_acc_name: encodedFromUsername, to_acc_username: encodedToUsername, amount: encodedAmount }),
    success: function(result) {
      // If the request is successful, update the account summary and transaction history
      //updateAccountSummary();
      //updateTransactionHistory();
      console.log("Transfer successful");
      window.location.reload();
    },
    error: function(xhr, status, error) {
      // If the request fails, display an error message
      alert("Error transferring funds: " + xhr.responseText);
      window.location.reload();
    }
  });
}
