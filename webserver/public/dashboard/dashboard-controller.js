'use strict';

angular.module('rc.dashboard', ['ngRoute'])

.config(['$routeProvider', function($routeProvider) {
  $routeProvider.when('/dashboard', {
    title: 'Settings',
    templateUrl: 'dashboard/dashboard.html',
    controller: 'DashboardCtrl'
  });
}])

.controller('DashboardCtrl', ['$scope', 'Config', 'ConfigClass', function($scope, Config, ConfigClass) {
  $scope.config = Config.get();
  $scope.classes = ConfigClass.query();

  $scope.saveConfig = function(config) {
    Config.save(config);
  };

  $scope.openAddClassPopup = function() {
    $scope.newClass = new ConfigClass();
    $('#addClassPopup').show();
  };

  $scope.addClass = function(newClass) {
    newClass.$save(function(data){
      $scope.classes.push(newClass);
    });
    $('#addClassPopup').hide();
  };

  $scope.deleteClass = function(delClass) {
    delClass.$delete(function() {
      $scope.classes = $.grep($scope.classes, function(e) {
        return e.id != delClass.id;
      });
    });
  };
}]);
