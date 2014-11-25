'use strict';

angular.module('rc.login', ['ngRoute'])

.config(['$routeProvider', function($routeProvider) {
  $routeProvider.when('/login', {
    templateUrl: 'login/login.html',
    controller: 'LoginCtrl'
  });
}])

.controller('LoginCtrl', ['$scope', '$rootScope', 'AUTH_EVENTS', 'AuthService',
  function($scope, $rootScope, AUTH_EVENTS, AuthService) {
  $scope.credentials = {
    email: '',
    password: ''
  };

  $scope.login = function(credentials) {
    AuthService.login(credentials).then(function(user){
      $rootScope.$broadcast(AUTH_EVENTS.loginSuccess);
    }, function() {
      $rootScope.$broadcast(AUTH_EVENTS.loginFailed);
    });
  };

  $scope.register = function(user) {

  };
}]);