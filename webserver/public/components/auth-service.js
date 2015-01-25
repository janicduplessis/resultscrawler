'use strict';

angular.module('rc.authservice', ['ngCookies'])

.factory('AuthService', ['$http', 'Session', function($http, Session) {
  var authService = {};

  authService.login = function(loginInfo) {
    return $http
      .post('/api/v1/auth/login', loginInfo)
      .then(function(res) {
        Session.create(res.data.token);
        return res.data.user;
      });
  };

  authService.register = function(registerInfo) {
    return $http
      .post('/api/v1/auth/register', registerInfo)
      .then(function(res) {
        Session.create(res.data.token);
        return res.data.user;
      });
  };

  authService.isAuthenticated = function() {
    return Session.authenticated;
  };

  return authService;
}])

.config(['$httpProvider', function($httpProvider) {
  $httpProvider.interceptors.push(['Session', function(Session) {
    return {
     request: function(config) {
        if(Session.authenticated) {
          config.headers['x-access-token'] = Session.token;
        }
        return config;
      }
    };
  }]);
}])

.service('Session', [function() {
  this.token = localStorage.getItem('token') || null;
  this.authenticated = !!this.token;
  this.create = function(token) {
    this.authenticated = true;
    this.token = token;
    localStorage.setItem('token', token);
  };
  this.destroy = function() {
    this.authenticated = false;
    this.token = null;
    localStorage.removeItem('token');
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
