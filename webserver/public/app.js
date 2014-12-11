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

.controller('ApplicationCtrl', ['$scope', '$route', '$location', '$mdSidenav', 'AuthService',
      function($scope, $route, $location, $mdSidenav, AuthService) {
  var modules = [],
      route,
      key;

  for(key in $route.routes) {
    route = $route.routes[key];
    if(route.menu) {
      modules.push({
        route: key,
        title: route.title,
        authentified: route.authentified,
        guest: route.guest
      });
    }
  }

  $scope.modules = modules;
  $scope.currentUser = null;

  $scope.menuClass = function(item) {
    $location.path().substring(1);
  };

  $scope.setCurrentUser = function(user) {
    $scope.currentUser = user;
  };

  $scope.toggleMenu = function() {
    $mdSidenav('left').toggle();
  };
}])

.run(['$location', '$rootScope', function($location, $rootScope) {
    $rootScope.$on('$routeChangeSuccess', function (event, current, previous) {
        $rootScope.title = current.$$route.title;
        // The controller name will be something like moduleNameCtrl,
        // to get the module name we remove Ctrl.
        var controllerName = current.$$route.controller;
        var moduleName = controllerName.toLowerCase().substring(0, controllerName.length - 4);
        $rootScope.moduleClass = 'rc-' + moduleName;
    });
}]);
