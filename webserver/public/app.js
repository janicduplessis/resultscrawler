'use strict';

// Declare app level module which depends on views, and components
angular.module('rc', [
  'ngRoute',
  'ngMaterial',
  'rc.sessions',
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

.controller('ApplicationCtrl', ['$rootScope', '$scope', '$route', '$location', '$mdSidenav', 'AuthService',
      function($rootScope, $scope, $route, $location, $mdSidenav, AuthService) {
  var modules = [],
      route,
      key;

  for(key in $route.routes) {
    route = $route.routes[key];
    if(route.menu) {
      modules.push({
        route: key,
        title: route.title,
        authentified: route.menu.authentified,
        guest: route.menu.guest,
        order: route.menu.order,
        selected: false
      });
    }
  }

  $rootScope.modules = modules;
  $scope.currentUser = AuthService.isAuthenticated();


  $scope.setCurrentUser = function(user) {
    $scope.currentUser = user;
  };

  $scope.displayMenuItem = function(item) {
    return $scope.currentUser && item.authentified || !$scope.currentUser && item.guest;
  };

  $scope.toggleMenuItem = function(item) {
    // Navigation
    $location.path(item.route);
    $mdSidenav('left').toggle();
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

        // Selection
        for(var i = 0; i < $rootScope.modules.length; i++) {
          $rootScope.modules[i].selected = $rootScope.modules[i].route === current.$$route.originalPath;
        }
    });
}]);
