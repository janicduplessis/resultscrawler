'use strict';

angular.module('rc.services', [])

.factory('AuthService', ['$http', 'Session', function($http, Session) {
  var authService = {};

  authService.login = function(credentials) {
    return $http
      .post('/api/v1/login', credentials)
      .then(function(res) {
        Session.create(res.data.user);
        return res.data.user;
      });
  };

  authService.isAuthenticated = function() {
    return !!Session.user;
  };

  return authService;
}])

.service('Session', function() {
  this.create = function(user) {
    this.user = user;
  };
  this.destroy = function() {
    this.user = null;
  };
  return this;
})

.constant('AUTH_EVENTS', {
  loginSuccess: 'auth-login-success',
  loginFailed: 'auth-login-failed',
  logoutSuccess: 'auth-logout-success',
  sessionTimeout: 'auth-session-timeout',
  notAuthenticated: 'auth-not-authenticated'
});