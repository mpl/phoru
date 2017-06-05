/*
Copyright 2017 Mathieu Lonjaret
*/
chrome.extension.onMessage.addListener(function(request, sender, sendResponse) {
	if (request.method == "getSelection") {
		sendResponse({data: window.getSelection().toString()});
		return;
	}
	if (request.method == "setSelection") {
		if (request.translation != null && request.translation != "") {
			// TODO(mpl): If possible, replace the selected text in the page with the
			// received translation. I don't think that's possible though.
			console.log("Received translation: " + request.translation);
		}
		sendResponse({});
		return;
	}		
	sendResponse({});
});
