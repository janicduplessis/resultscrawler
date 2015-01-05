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

.controller('ResultsCtrl', ['$scope', '$timeout', '$mdDialog', 'Results', 'Sessions', function($scope, $timeout, $mdDialog, Results, Sessions) {
  $scope.session = Sessions.getUserCurrent();
  $scope.results = Results.get({year: $scope.session});

  $scope.changeYear = function(year) {
    $scope.results = Results.get(year);
  };

  $scope.openChangeSessionDialog = function(ev) {
    $mdDialog.show({
      controller: 'ChangeSessionDialogCtrl',
      templateUrl: 'results/choose-session-dialog.tmpl.html',
      targetEvent: ev
    })
    .then(function(answer) {
      Sessions.setUserCurrent(answer);
      $scope.session = answer;
      $scope.results = Results.get({year: $scope.session});
    });
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
}])

.controller('ChangeSessionDialogCtrl', ['$scope', '$mdDialog', 'Sessions', function($scope, $mdDialog, Sessions) {
  var sessions = Sessions.list();
  var curSession = Sessions.getUserCurrent();
  for(var i = 0; i < sessions.length; i++) {
    if(sessions[i].value === curSession) {
      sessions[i].selected = true;
    } else {
      sessions[i].selected = false;
    }
  }

  $scope.sessions = sessions;

  $scope.hide = function() {
    $mdDialog.hide();
  };
  $scope.cancel = function() {
    $mdDialog.cancel();
  };
  $scope.changeSession = function(answer) {
    $mdDialog.hide(answer);
  };
}]);
