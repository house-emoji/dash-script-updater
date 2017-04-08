"use strict";

window.onload = function() {
	var updateButton = document.getElementById("update-button");
	
	updateButton.onclick = update
}

function update() {
	var xhttp = new XMLHttpRequest();
	xhttp.onreadystatechange = function() {
		if (this.readyState == 4) {
			document.getElementById("console").innerHTML = this.responseText;

			if (this.status == 200) {
				document.getElementById("status-text").innerHTML = "Success!";
			} else {
				document.getElementById("status-text").innerHTML = this.status + ": Failed to update";
			}
		}
	}
	xhttp.open("GET", "/update", true);
	xhttp.send();
}
