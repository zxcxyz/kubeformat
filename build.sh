tag=1.0.3
go build -o ./build/kubeformat-darwin-amd64-${tag} github.com/zxcxyz/kubeformat
env GOOS=linux GOARCH=amd64 go build -o ./build/kubeformat-linux-amd64-${tag} github.com/zxcxyz/kubeformat