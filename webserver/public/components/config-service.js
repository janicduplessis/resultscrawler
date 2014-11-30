'use strict';

angular.module('rc.configservice', ['ngResource'])

.factory('Config', ['$resource', function($resource) {
  return $resource('/api/v1/crawler/config');
}])

.factory('ConfigClass', ['$resource', function($resource) {
  return $resource('/api/v1/crawler/class/:id', {
      id: '@id'
    }, {
      update: {
        method: 'PUT'
    }
  });
}]);
