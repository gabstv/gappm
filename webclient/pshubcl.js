/*global jQuery:false */
/*global console:false */
"use strict";

jQuery.support.cors = true;
var pshubcl = {};
pshubcl.appid = "";
pshubcl.clientkey = "";
pshubcl.clientid = "";
pshubcl.host = "";
pshubcl._listen = true;

pshubcl._post = function(sender, path, callback){
	//console.log(JSON.stringify(sender));
	jQuery.ajax({
	  url: pshubcl.host + path,
	  type: "POST",
	  dataType: "json",
	  data: JSON.stringify(sender),
	  success: function(data, textStatus, xhr) {
	    //called when successful
	    //console.log(xhr);
	    //console.log("it kinda worked");
	    callback(true, "", data);
	  },
	  error: function(xhr, textStatus, errorThrown) {
	    //called when there is an error
	    //console.log("it didn't work");
	    //console.log(textStatus);
	    //console.log(xhr);
	    //console.log(xhr.status);
	    callback(false, errorThrown, null);
	  },
	  contentType: "application/json",
	  accepts: "application/json",
	  cache: false,
	  crossDomain: true
	});
	
};

pshubcl._new_request = function(message){
	var outp = {
		app_id: pshubcl.appid,
		key: pshubcl.clientkey,
		message: message
	};
	return outp;
};

pshubcl.connect = function(url, app, callback){
	pshubcl.host = url;
	pshubcl.appid = app;
	var rq = pshubcl._new_request();
	pshubcl._post(rq, "/connect", function(success, errormsg, data){
		if(!success){
			callback(false, errormsg);
			return;
		}
		pshubcl.clientkey = data.key;
		pshubcl.clientid = data.id;
		callback(true, "");
	});
};

pshubcl.subscribe = function(channel, password, callback){
	var str0 = channel;
	if(password !== undefined){
		if(password.length > 0){
			str0 = channel + ":" + password;
		}
	}
	var rq = pshubcl._new_request();
	pshubcl._post(rq, "/subscribe/"+str0, function(success, errormsg, data){
		if(!success){
			callback(false, errormsg);
			return;
		}
		callback(data.success, data.info);
	});
};

pshubcl.startlisten = function(callback){
	pshubcl._listen = true;
	pshubcl.__listen(callback);
};

pshubcl.stoplisten = function(){
	pshubcl._listen = false;
};

pshubcl.__listen = function(callback){
	var rq = pshubcl._new_request();
	pshubcl._post(rq, "/listen", function(success, errormsg, data){
		if(!success){
			callback(false, errormsg);
		} else if(!data.success){
			callback(false, data.info);
		} else {
			for(var i = 0; i < data.messages.length; i++){
				callback(true, "", data.messages[i]);
			}
		}
		if(pshubcl._listen){
			pshubcl.__listen(callback);
		}
	});
};