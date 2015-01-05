'use strict';

angular.module('rc.results', ['ngRoute'])

.config(['$routeProvider', function($routeProvider) {
  $routeProvider.when('/results', {
  	title: 'Results',
    templateUrl: 'results/results.html',
    controller: 'ResultsCtrl',
    menu: {
      authentified: true,
      guest: false,
      order: 3
    }
  });
}])

.controller('ResultsCtrl', ['$scope', '$timeout', 'Results', function($scope, $timeout, Results) {
  $scope.session = '20143';
  $scope.results = Results.get({year: $scope.session});

  $scope.changeYear = function(year) {
    $scope.results = Results.get(year);
  };

  $scope.changeSession = function() {

  };

  $scope.refresh = function() {
    Results.refresh().success(function() {
      // hacky hacky wait 1 sec for the crawler to complete its run.
      // should have a way to ping server for progress...
      $timeout(function() {
         $scope.results = Results.get({year: $scope.session});
      }, 1000);

    }).error(function() {
      //TODO: handle errors here
    });
  };
}]);
