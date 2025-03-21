# raft-assignment

go mod init github.com/MuhammadTaimoorAnwar511/raft-assignment
go mod tidy

.\run_nodes.bat
working fine till now leader election

or

go run cmd/main.go --id=1 --port=9001 --peers=localhost:9002,localhost:9003

go run cmd/main.go --id=2 --port=9002 --peers=localhost:9001,localhost:9003

go run cmd/main.go --id=3 --port=9003 --peers=localhost:9001,localhost:9002

put myKey hello

get myKey

