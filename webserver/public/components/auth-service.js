'use strict';

angular.module('rc.authservice', ['ngCookies'])

.factory('AuthService', ['$http', 'Session', function($http, Session) {
  var authService = {};

  authService.login = function(loginInfo) {
    return $http
      .post('/api/v1/auth/login', loginInfo)
      .then(function(res) {
        Session.create();
        return res.data.user;
      });
  };

  authService.register = function(registerInfo) {
    return $http
      .post('/api/v1/auth/register', registerInfo)
      .then(function(res) {
        Session.create();
        return res.data.user;
      });
  };

  authService.isAuthenticated = function() {
    return Session.authenticated;
  };

  return authService;
}])

.service('Session', ['$cookies', function($cookies) {
  this.authenticated = $cookies.authenticated === 'true';
  this.create = function() {
    this.authenticated = true;
    $cookies.authenticated = 'true';
  };
  this.destroy = function() {
    this.authenticated = false;
    delete $cookies.authenticated;
  };
  return this;
}])

.constant('AUTH_EVENTS', {
  loginSuccess: 'auth-login-success',
  loginFailed: 'auth-login-failed',
  logoutSuccess: 'auth-logout-success',
  sessionTimeout: 'auth-session-timeout',
  notAuthenticated: 'auth-not-authenticated'
});
