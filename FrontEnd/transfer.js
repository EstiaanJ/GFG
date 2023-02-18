console.log("transfer.js loaded");

function transfer() {
    var to_acc_username = document.getElementById("to_acc_username").value;
    var amount = document.getElementById("amount").value;
    var from_acc_name = document.getElementById("username").textContent;

    console.log("Transfering " + amount + " from " + from_acc_name + " to " + to_acc_username);
    
    // Make a POST request to the /transfer endpoint with the from_acc_username and amount values
    $.ajax({
      url: "http://51.161.163.66:44658/transfer",
      type: "POST",
      dataType: "json",
      data: JSON.stringify({from_acc_name:from_acc_name,to_acc_username:to_acc_username, amount: amount }),
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
  