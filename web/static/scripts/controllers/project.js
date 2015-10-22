"use strict";angular.module("decapApp").controller("ProjectController",["$scope","$stateParams","$uibModal","$log","DecapService",function($scope,$stateParams,$uibModal,$log,DecapService){$scope.projectName=$stateParams.project,DecapService.getProject($scope.projectName).then(function(project){$scope.project=project,DecapService.getBuilds($scope.project.team,$scope.project.project).then(function(builds){for(var b in builds)builds[b].timestamp=new Date(1e3*builds[b].startTime).toISOString();$scope.builds=builds},function(message){console.log("error getting builds list: "+message)}),DecapService.getBranches($scope.project.team,$scope.project.project).then(function(refs){$scope.refs=refs},function(message){console.log("error getting builds list: "+message)})},function(message){console.log("error getting projects list: "+message)}),$scope.animationsEnabled=!0,$scope.viewConsole=function(buildId){DecapService.getLogs(buildId).then(function(output){var downloadURL=DecapService.getLogsDownloadURL(buildId),modalInstance=$uibModal.open({animation:$scope.animationsEnabled,templateUrl:"modalConsole.html",controller:"ConsoleController",size:"lg",resolve:{id:function(){return buildId},output:function(){return output},downloadURL:function(){return downloadURL}}});modalInstance.result.then(function(selectedItem){$scope.selected=selectedItem},function(){$log.info("Modal dismissed at: "+new Date)})},function(message){console.log("error getting builds list: "+message)})},$scope.viewArtifacts=function(buildId){DecapService.getArtifacts(buildId).then(function(output){var downloadURL=DecapService.getArtifactsDownloadURL(buildId),modalInstance=$uibModal.open({animation:$scope.animationsEnabled,templateUrl:"modalArtifacts.html",controller:"ArtifactsController",size:"lg",resolve:{id:function(){return buildId},output:function(){return output},downloadURL:function(){return downloadURL}}});modalInstance.result.then(function(selectedItem){$scope.selected=selectedItem},function(){$log.info("Modal dismissed at: "+new Date)})},function(message){console.log("error getting builds list: "+message)})}}]),angular.module("decapApp").controller("ConsoleController",function($scope,$modalInstance,id,output,downloadURL){$scope.id=id,$scope.output=output,$scope.downloadURL=downloadURL,$scope.ok=function(){$modalInstance.close()},$scope.cancel=function(){$modalInstance.dismiss("cancel")}}),angular.module("decapApp").controller("ArtifactsController",function($scope,$modalInstance,id,output,downloadURL){$scope.id=id,$scope.output=output,$scope.downloadURL=downloadURL,$scope.ok=function(){$modalInstance.close()},$scope.cancel=function(){$modalInstance.dismiss("cancel")}});