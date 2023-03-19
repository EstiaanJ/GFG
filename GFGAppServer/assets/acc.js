$(document).ready(function() {
    // Read query parameters from URL
    var queryString = window.location.search;
    var urlParams = new URLSearchParams(queryString);
    var username = urlParams.get("username");
    var hash = urlParams.get("hash");
    
    // Make AJAX request to server
    $.ajax({
      type: "POST",
      url: "http://localhost:44658" + hash,
      dataType: "json",
      data: JSON.stringify({username: username}), // Send only username in JSON object
      success: function(data) {
        // Update account summary section
        $("#account-number").text(data.account_number);
        $("#balance").text("℣" + data.balance.toFixed(3));
        //$("#last-transaction").text(data.last_transaction);
        $("#username").text(username);
        
        // Update transaction history table
        var transactionTable = $("#transaction-table");
        data.transactions.forEach(function(transaction) {
          var row = $("<tr></tr>");
          
          // Format date string as yyyy-mm-dd
          var date = new Date(transaction.date).toISOString().slice(0, 10);
          
          row.append($("<td>" + date + "</td>"));
          row.append($("<td>" + transaction.game_date + "</td>"));
          row.append($("<td>" + "℣" + transaction.amount.toFixed(3) + "</td>"));
          //Desecription: Limit to 1000 characters
          transaction.description = transaction.description.substring(0, 1000);
          row.append($("<td>" + transaction.description + "</td>"));
          transactionTable.append(row);
        });
        
      },
      error: function() {
        alert("Failed to fetch account details");
      }
    });
  });
  

