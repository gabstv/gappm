/*global jQuery:false */
/*global console:false */
/*global $:false */
/*global pshubcl:false */
"use strict";

//function rand(a, b) {
//	var diff = a;
//	if (b !== undefined) {
//		diff = b - a;
//	} else {
//		b = a;
//		a = 0;
//	}
//	return Math.floor(Math.random() * diff) + a;
//}
//function makeid()
//{
//    var text = "";
//    var possible = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789";
//    var max = rand(10,60);
//    for( var i=0; i < max; i++ )
//        text += possible.charAt(rand(possible.length));
//
//    return text;
//}
//function color()
//{
//	var text = "";
//    var possible = ["red", "#FFF510", "#2121FF", "#FE16C9", "#CCC", "#FFFFFE"];
//    text = possible[rand(possible.length)];
//    return text;
//}
//function test () {
//	$("#console").append('<span style="color:' + color() + ';">' + makeid() + '</span> \n')
//	var console    = $('#console');
//	var height = console[0].scrollHeight;
//	console.scrollTop(height);
//}
function onmessage(success, errmsg, jmsg){
	//console.log(errmsg);
	if(jmsg !== undefined){
		console.log(jmsg);
		$("#console").append(jmsg.message);
		var console2    = $("#console");
		var height = console2[0].scrollHeight;
		console2.scrollTop(height);
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
jQuery(document).ready(function() {
	//setInterval(test, 1000)
	connect();
});