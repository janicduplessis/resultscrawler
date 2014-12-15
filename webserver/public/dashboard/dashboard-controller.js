'use strict';

angular.module('rc.dashboard', ['ngRoute'])

.config(['$routeProvider', function($routeProvider) {
  $routeProvider.when('/dashboard', {
    title: 'Setup',
    templateUrl: 'dashboard/dashboard.html',
    controller: 'DashboardCtrl',
    menu: {
      authentified: true,
      guest: false,
      order: 5
    }
  });
}])

.controller('DashboardCtrl', ['$scope', '$mdDialog', 'Config', 'ConfigClass',
    function($scope, $mdDialog, Config, ConfigClass) {

  $scope.config = Config.get();
  $scope.classes = ConfigClass.query();

  $scope.saveConfig = function(config) {
    Config.save(config);
  };

  $scope.openAddClassPopup = function(ev) {

    $mdDialog.show({
      controller: AddClassDialogCtrl,
      templateUrl: 'dashboard/add-class-dialog.tmpl.html',
      targetEvent: ev
    })
    .then(function(answer) {
      var newClass = new ConfigClass();
      newClass.name = answer.name;
      newClass.group = answer.group;
      newClass.year = answer.year;
      newClass.$save();
      $scope.classes.push(newClass);
    });
  };

  $scope.deleteClass = function(delClass) {
    delClass.$delete(function() {
      $scope.classes = $.grep($scope.classes, function(e) {
        return e.id != delClass.id;
      });
    });
  };
}]);

function AddClassDialogCtrl($scope, $mdDialog) {
  $scope.hide = function() {
    $mdDialog.hide();
  };
  $scope.cancel = function() {
    $mdDialog.cancel();
  };
  $scope.answer = function(answer) {
    $mdDialog.hide(answer);
  };
}
