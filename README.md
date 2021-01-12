# kubeformat
Tool to remove junk from kubectl manifests. For those of you who are also tired of kubectl flooding you with useless information. 
### Features:
* Cleans up fields using filters defined in cmd/defaults.go
* Iterates over containers(container filters are defined with \*, "spec.template.spec.containers.\*.terminationMessagePath")
* Removes empty fields
### Example
![example](./example.png)
### Upcoming features
* Output in JSON
* Read input from files
* Importing custom filters 
# Usage
```sh
kubectl -n default get deployment | kubeformat
```
