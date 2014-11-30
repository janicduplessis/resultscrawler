'use strict';

angular.module('rc.resultsservice', ['ngResource'])

.factory('Results', ['$resource', function($resource) {
  return $resource('/api/v1/results/:year');
}]);
