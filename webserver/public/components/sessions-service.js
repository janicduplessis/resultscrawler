'use strict';

angular.module('rc.sessions', ['ngCookies'])

.factory('Sessions', ['$filter', 'SessionStorage', function($filter, SessionStorage) {
	return {
		// getCurrent returns the current session based on the current date.
		getCurrent: function() {
			var curDate = new Date();
			var month = curDate.getMonth();
			var s = curDate.getFullYear().toString();
			if(month >= 0 && month < 6) {
				// winter
				s += '1';
			} else if(month >= 6 && month < 8) {
				// summer
				s += '2';
			} else {
				// fall
				s += '3';
			}
			return s;
		},

		// getUserCurrent gets the session the user selected.
		getUserCurrent: function() {
			var s = SessionStorage.get();
			if(s) {
				return s;
			}

			s = this.getCurrent();
			SessionStorage.set(s);
			return s;
		},

		// setUserCurrent sets the current session for the user.
		setUserCurrent: function(session) {
			SessionStorage.set(session);
		},

		// list returns available sessions.
		// TODO: maybe move this server side so we can keep old sessions
		// that have results.
		list: function() {
			// Cache the result of this function.
			if(this.list.data) {
				return this.list.data;
			}
			var list = [];
			var cur = this.getCurrent();
			var num = 9;
			var curYear = parseInt(cur.substring(0, 4), 10);
			var curSession = parseInt(cur.charAt(4), 10);
			while(list.length <= num) {
				cur = curYear.toString() + curSession.toString();
				list.push({
					value: cur,
					name: $filter('session')(cur)
				});

				if(curSession > 1) {
					curSession--;
				} else {
					curYear--;
					curSession = 3;
				}
			}
			this.list.data = list;
			return this.list.data;
		}
	};
}])

.service('SessionStorage', ['$cookies', function($cookies) {
  this.currentSession = $cookies.currentSession;
  this.get = function() {
    return this.currentSession;
  };
  this.set = function(session) {
    this.currentSession = session;
    $cookies.currentSession = session;
  };
  return this;
}])

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
