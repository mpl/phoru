/*
Copyright 2017 Mathieu Lonjaret
*/

chrome.runtime.onMessage.addListener(function(request, sender, sendResponse) {
	if (request.method != "getSelection") {
		return;
	}
	var selection = googleDocsUtil.getGoogleDocument().selectedText;
	if (!selection || selection == "") {
		return;
	}
	sendResponse({data: selection});
	console.log("Sending selection: " + selection);
});
