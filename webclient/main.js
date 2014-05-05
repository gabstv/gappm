/*global jQuery:false */
/*global console:false */
/*global $:false */
/*global pshubcl:false */
"use strict";

var fc = 0;

function onmessage(success, errmsg, jmsg){
	//console.log(errmsg);
	if(jmsg !== undefined){
		fc++;
		//console.log(jmsg);
		$("#console").append(jmsg.message);
		var console2    = $("#console");
		var height = console2[0].scrollHeight;
		if($("#olaybox").is(":checked"))
			console2.scrollTop(height);
		if(fc > 600){
			cleanup();
		}
	}
}
function connect() {
	var hash = window.location.hash.substring(1);
	pshubcl.connect("http://" + hash + ":5996", "2222be5d-8491-44e1-b3f1-b1528b37fe94", function(success, errmsg){
		if(!success){
			console.log("CONNECTION ERROR!");
			console.log(errmsg);
			connect();
			return;
		}
		pshubcl.subscribe("gappm", "", function(success, errmsg){
			if(!success){
				console.log("SUBSCRIBE ERROR!");
				console.log(errmsg);
				//connect();
				return;
			}
			pshubcl.startlisten(onmessage);
		});
	});
}
function cleanup() {
	console.log("CLEANUP!!!!!");
	if(fc < 500)
		return;
	var h = $("#console").html().split("<br>").slice(100);
	fc = h.length;
	$("#console").html(h.join("<br>"));
}
jQuery(document).ready(function() {
	connect();
});