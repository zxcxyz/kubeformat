/*
Copyright © 2021 zxcxyz <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

var defaultFilters = `[
    "metadata.managedFields",
    "metadata.labels",
    "metadata.annotations.kubectl\\.kubernetes\\.io/last-applied-configuration",
    "metadata.annotations.deployment\\.kubernetes\\.io/revision",
    "metadata.creationTimestamp",
    "metadata.generation",
    "metadata.resourceVersion",
    "metadata.selfLink",
    "metadata.uid",
    "spec.revisionHistoryLimit",
    "spec.template.metadata.creationTimestamp",
    "spec.progressDeadlineSeconds",
    "spec.strategy",
    "spec.template.spec.restartPolicy",
    "spec.template.spec.dnsPolicy",
    "status",
    "spec.template.spec.terminationGracePeriodSeconds",
    "spec.template.spec.schedulerName"
]`
var containerFilters = `[
    "spec.template.spec.containers.*.livenessProbe.failureThreshold",
    "spec.template.spec.containers.*.livenessProbe.initialDelaySeconds",
    "spec.template.spec.containers.*.livenessProbe.periodSeconds",
    "spec.template.spec.containers.*.livenessProbe.successThreshold" ,
    "spec.template.spec.containers.*.livenessProbe.timeoutSeconds",
    "spec.template.spec.containers.*.imagePullPolicy",
    "spec.template.spec.containers.*.terminationMessagePath",
    "spec.template.spec.containers.*.terminationMessagePolicy"
]`
var emptyCheckFilters = `[
    "metadata.annotations",
    "spec.template.spec.securityContext"
]`

// "metadata.annotations."kubernetes.io/service-account.name""
// "metadata.annotations."kubernetes.io/service-account.uid""
