'use strict';

angular.module('rc.about', ['ngRoute'])

.config(['$routeProvider', function($routeProvider) {
  $routeProvider.when('/about', {
  	title: 'About',
    templateUrl: 'about/about.html',
    controller: 'AboutCtrl',
    menu: {
      authentified: true,
      guest: true,
      order: 10
    }
  });
}])

.controller('AboutCtrl', [function() {

}]);
