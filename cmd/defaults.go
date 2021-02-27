/*
Package cmd blah

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

var defaultFilters = `{
    "filters": 
    [
    "metadata.managedFields",
    "metadata.annotations.kubectl\\.kubernetes\\.io/last-applied-configuration",
    "metadata.annotations.deployment\\.kubernetes\\.io/revision",
    "metadata.creationTimestamp",
    "metadata.generation",
    "metadata.resourceVersion",
    "metadata.selfLink",
    "metadata.uid",
    "metadata.ownerReferences",
    "metadata.finalizers",
    "metadata.namespace",
    "spec.template.metadata.creationTimestamp",
    "spec.progressDeadlineSeconds",
    "spec.clusterIP",
    "status",
    "spec.template.spec.terminationGracePeriodSeconds",
    "spec.template.spec.containers.*.terminationMessagePath",
    "spec.template.spec.containers.*.terminationMessagePolicy"
]}`

// items[*].spec.volumeMode
// items[*].spec.volumeName
// items[*].spec.volumeClaimTemplates.metadata.creationTimestamp.
// "metadata.namespace"
// "metadata.annotations."kubernetes.io/service-account.name""
// "metadata.annotations."kubernetes.io/service-account.uid""
