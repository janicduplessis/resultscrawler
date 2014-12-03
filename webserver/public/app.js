'use strict';

// Declare app level module which depends on views, and components
angular.module('rc', [
  'ngRoute',
  'ngMaterial',
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
  //$locationProvider.html5Mode(true).hashPrefix('!');
}])

.controller('ApplicationCtrl', ['$scope', '$mdSidenav', 'AuthService', function($scope, $mdSidenav, AuthService) {
  $scope.currentUser = null;

  $scope.setCurrentUser = function(user) {
    $scope.currentUser = user;
  };

  $scope.openMenu = function() {
    $mdSidenav('left').toggle();
  };
}])

.run(['$location', '$rootScope', function($location, $rootScope) {
    $rootScope.$on('$routeChangeSuccess', function (event, current, previous) {
        $rootScope.title = current.$$route.title;
    });
}]);
