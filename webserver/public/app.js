'use strict';

// Declare app level module which depends on views, and components
angular.module('rc', [
  'ngRoute',
  'rc.authservice',
  'rc.configservice',
  'rc.resultsservice',
  'rc.home',
  'rc.about',
  'rc.login',
  'rc.dashboard',
  'rc.results'
]).
config(['$routeProvider', '$locationProvider', function($routeProvider, $locationProvider) {
  $routeProvider.otherwise({redirectTo: '/home'});
  $locationProvider.html5Mode(true).hashPrefix('!');
}])

.controller('ApplicationCtrl', ['$scope', 'AuthService', function($scope, AuthService) {
  $scope.currentUser = null;

  $scope.setCurrentUser = function(user) {
    $scope.currentUser = user;
  };
}]);
