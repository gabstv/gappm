"use strict";
var math2 = {};
math2.rand = function(a, b) {
	var diff = a;
	if (b !== undefined) {
		diff = b - a;
	} else {
		b = a;
		a = 0;
	}
	return Math.floor(Math.random() * diff) + a;
};