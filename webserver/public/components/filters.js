'use strict';

angular.module('rc.filters', [])

.filter('session', function() {
	var sessions = [
		'Winter',
		'Summer',
		'Fall'
	];

	return function(input) {
		if(typeof input !== 'string' || input.length !== 5) {
			return '';
		}
		var session = sessions[parseInt(input.charAt(4)) - 1];
		return session + ' ' + input.substring(0,4);
	};
});
