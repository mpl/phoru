/*
Copyright 2017 Mathieu Lonjaret
*/

function renderStatus(statusText) {
	document.getElementById('status').textContent = statusText;
}

document.addEventListener('DOMContentLoaded', function() {
	var t = setTimeout(function(){
		var result = document.getElementById('result');
		if (!result || !result.textContent || result.textContent == "") {
			renderStatus("No selection. Highlight some text with your mouse cursor.");
		}
	},1000);
	translateSelection(t);
});

function translateSelection(timerId) {
	chrome.tabs.query({active:true, currentWindow: true},
	function(tabs) {
		chrome.tabs.sendMessage(tabs[0].id, {method: "getSelection"}, 
		function(response){
			if (response == null || response.data == "") {
				return;
			}
			clearTimeout(timerId);
			phoru(response.data);
		});
	});
}

function phoru(selection) {
	var translated = gophoru.Run(selection, function(errMsg) {
		renderStatus("Error: " + errMsg);
	});
	if (translated == "") {
		return;		
	}
	var result = document.getElementById('result');
	result.textContent = translated;
	// TODO(mpl): do it with CSS or whatever.
	result.rows = 30;
	result.cols = 50;
	result.hidden = false;
}
