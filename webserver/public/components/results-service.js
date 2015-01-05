'use strict';

angular.module('rc.resultsservice', ['ngResource'])

.factory('Results', ['$http', '$resource', function($http, $resource) {
  var resource = $resource('/api/v1/results/:year');
  resource.refresh = function() {
  	return $http.post('/api/v1/crawler/refresh');
  };
  return resource;
}]);
