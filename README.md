# kubeformat
Tool to remove junk from kubectl manifests. For those of you who are also tired of kubectl flooding you with useless information. Also my coursework. Pretty much the same as kubectl-neat, it was just too late for me to change coursework theme when I found it out.
### Features:
* Cleans up fields using filters defined in cmd/defaults.go
* Iterates over containers(container filters are defined with \*, "spec.template.spec.containers.\*.terminationMessagePath")
* Removes empty fields
* Able to import custom filters. Sample JSON with filters can be found in sample/. Usage: `... | kubeformat -p filepath`. Note that wildcard (*) only works for containers, and you should escape irrelevant dots with \\
# Example
![example](./example.png)
### Upcoming features
* Output in JSON
* Read input from files
* ~~Importing custom filters~~
* Maybe add statefulness (path to filters)
# Usage
```sh
kubectl -n default get deployment | kubeformat
```
