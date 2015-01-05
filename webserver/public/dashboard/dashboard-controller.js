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

.controller('DashboardCtrl', ['$scope', '$mdDialog', 'Config', 'ConfigClass', 'Sessions',
    function($scope, $mdDialog, Config, ConfigClass, Sessions) {

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
      newClass.year = answer.year.value;
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

  // Not sure how dependency injection works here... Could be cleaner.
  var sessions = Sessions.list();
  function AddClassDialogCtrl($scope, $mdDialog) {
    $scope.sessions = sessions;

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

}]);
