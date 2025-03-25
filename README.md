# raft-assignment

go mod init github.com/MuhammadTaimoorAnwar511/raft-assignment
go mod tidy
===
# (optional RUN THESE 3 SEPERATELY OR .\run_nodes.bat)
go run cmd/main.go --id=1 --port=9001 --peers=localhost:9002,localhost:9003

go run cmd/main.go --id=2 --port=9002 --peers=localhost:9001,localhost:9003

go run cmd/main.go --id=3 --port=9003 --peers=localhost:9001,localhost:9002

# OR

# run this command on vs code terminal
.\run_nodes.bat

# run these commands on leader node:

put myKey hello

get myKey

append myKey world

exit

# Here’s the correct full set for 4 peers:
go run cmd/main.go --id=1 --port=9001 --peers=localhost:9002,localhost:9003,localhost:9004
go run cmd/main.go --id=2 --port=9002 --peers=localhost:9001,localhost:9003,localhost:9004
go run cmd/main.go --id=3 --port=9003 --peers=localhost:9001,localhost:9002,localhost:9004
go run cmd/main.go --id=4 --port=9004 --peers=localhost:9001,localhost:9002,localhost:9003

# Here’s the correct full set for 5 peers:
go run cmd/main.go --id=1 --port=9001 --peers=localhost:9002,localhost:9003,localhost:9004,localhost:9005
go run cmd/main.go --id=2 --port=9002 --peers=localhost:9001,localhost:9003,localhost:9004,localhost:9005
go run cmd/main.go --id=3 --port=9003 --peers=localhost:9001,localhost:9002,localhost:9004,localhost:9005
go run cmd/main.go --id=4 --port=9004 --peers=localhost:9001,localhost:9002,localhost:9003,localhost:9005
go run cmd/main.go --id=5 --port=9005 --peers=localhost:9001,localhost:9002,localhost:9003,localhost:9004
