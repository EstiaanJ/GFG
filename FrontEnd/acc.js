$(document).ready(function() {
    // Read query parameters from URL
    var queryString = window.location.search;
    var urlParams = new URLSearchParams(queryString);
    var username = urlParams.get("username");
    var hash = urlParams.get("hash");
    
    // Make AJAX request to server
    $.ajax({
      type: "POST",
      url: "http://localhost:8384" + hash,
      dataType: "json",
      data: JSON.stringify({username: username}), // Send only username in JSON object
      success: function(data) {
        // Update account summary section
        $("#account-number").text(data.account_number);
        $("#balance").text("$" + data.balance.toFixed(2));
        $("#last-transaction").text(data.last_transaction);
        $("#username").text(username);
        
        // Update transaction history table
        var transactionTable = $("#transaction-table");
        data.transaction_history.forEach(function(transaction) {
          var row = $("<tr></tr>");
          row.append($("<td>" + transaction.date + "</td>"));
          row.append($("<td>" + transaction.description + "</td>"));
          row.append($("<td>" + "$" + transaction.amount.toFixed(2) + "</td>"));
          transactionTable.append(row);
        });
      },
      error: function() {
        alert("Failed to fetch account details");
      }
    });
  });
  

