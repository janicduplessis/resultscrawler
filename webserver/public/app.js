'use strict';

// Declare app level module which depends on views, and components
angular.module('rc', [
  'ngRoute',
  'rc.authservice',
  'rc.configservice',
  'rc.home',
  'rc.about',
  'rc.login',
  'rc.dashboard'
]).
config(['$routeProvider', function($routeProvider) {
  $routeProvider.otherwise({redirectTo: '/home'});
}]);
