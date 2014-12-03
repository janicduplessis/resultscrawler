'use strict';

angular.module('rc.login', ['ngRoute'])

.config(['$routeProvider', function($routeProvider) {
  $routeProvider.when('/login', {
    title: 'Sign in',
    templateUrl: 'login/login.html',
    controller: 'LoginCtrl'
  }).when('/register', {
    title: 'Register',
    templateUrl: 'login/register.html',
    controller: 'LoginCtrl'
  }).when('/logout', {
    controller: 'LogoutCtrl'
  });
}])

.controller('LoginCtrl', ['$scope', '$rootScope', 'AUTH_EVENTS', 'AuthService',
      function($scope, $rootScope, AUTH_EVENTS, AuthService) {

  $scope.loginInfo = {
    email: '',
    password: ''
  };

  $scope.login = function(loginInfo) {
    AuthService.login(loginInfo).then(function(user){
      $rootScope.$broadcast(AUTH_EVENTS.loginSuccess);
      $scope.setCurrentUser(user);
    }, function() {
      $rootScope.$broadcast(AUTH_EVENTS.loginFailed);
    });
  };
}])

.controller('RegisterCtrl', ['$scope', '$rootScope', 'AUTH_EVENTS', 'AuthService',
      function($scope, $rootScope, AUTH_EVENTS, AuthService) {

  $scope.registerInfo = {
    email: '',
    password: '',
    firstName: '',
    lastName: ''
  };

  $scope.register = function(registerInfo) {
    AuthService.register(registerInfo).then(function(user){
      $rootScope.$broadcast(AUTH_EVENTS.loginSuccess);
    }, function() {
      $rootScope.$broadcast(AUTH_EVENTS.loginFailed);
    });
  };
}]).controller('LogoutCtrl', ['$scope', '$rootScope', '$location', 'AUTH_EVENTS', 'AuthService',
  function($scope, $rootScope, $location, AUTH_EVENTS, AuthService) {
      AuthService.logout();
      $rootScope.$broadcast(AUTH_EVENTS.logoutSuccess);
      $scope.setCurrentUser(null);
      $location.path('/home');
}]);
