document.getElementById("submitButton").addEventListener("click", function() {
    var inputText = document.getElementById("inputText").value;
    fetch('http://127.0.0.1:1000/', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({text: inputText}), 
    })
    .then(response => response.json())
    .then(data => console.log(data))
    .catch((error) => console.error('Error:', error));
});
