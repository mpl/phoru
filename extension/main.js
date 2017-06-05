/*
Copyright 2017 Mathieu Lonjaret
*/

function renderStatus(statusText) {
	document.getElementById('status').textContent = statusText;
}

document.addEventListener('DOMContentLoaded', function() {
	chrome.tabs.query({active:true, currentWindow: true}, 
	function(tabs) {
		chrome.tabs.sendMessage(tabs[0].id, {method: "getSelection"}, 
		function(response){
			if (response == null || response.data == "") {
				renderStatus("No selection. Highlight some text with your mouse cursor.");
				return;
			}
			phoru(response.data);
		});
	});
});

function phoru(selection) {
	var translated = gophoru.Run(selection, function(errMsg) {
		renderStatus("Error: " + errMsg);
	});
	if (translated == "") {
		// TODO(mpl): Can this ever happen? nm for now.
		renderStatus("No selection. Highlight some text with your mouse cursor.");
		return;		
	}
	var result = document.getElementById('result');
	result.textContent = translated;
	result.hidden = false;

	chrome.tabs.query({active:true, currentWindow: true}, 
	function(tabs) {
		chrome.tabs.sendMessage(tabs[0].id, {method: "setSelection", translation: translated});
	});
}
