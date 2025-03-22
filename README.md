# raft-assignment

go mod init github.com/MuhammadTaimoorAnwar511/raft-assignment
go mod tidy
===

go run cmd/main.go --id=1 --port=9001 --peers=localhost:9002,localhost:9003

go run cmd/main.go --id=2 --port=9002 --peers=localhost:9001,localhost:9003

go run cmd/main.go --id=3 --port=9003 --peers=localhost:9001,localhost:9002

or
# run this command on vs code terminal
.\run_nodes.bat

# run these commands on leader node:

put myKey hello

get myKey

