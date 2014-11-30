'use strict';

angular.module('rc.results', ['ngRoute'])

.config(['$routeProvider', function($routeProvider) {
  $routeProvider.when('/results', {
    templateUrl: 'results/results.html',
    controller: 'ResultsCtrl'
  });
}])

.controller('ResultsCtrl', ['$scope', 'Results', function($scope, Results) {
  $scope.results = Results.get({year: '20143'});

  $scope.changeYear = function(year) {
    $scope.results = Results.get(year);
  };
}]);
