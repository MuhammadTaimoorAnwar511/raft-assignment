@echo off
echo Starting Raft Node 1...
start cmd /k "cd /d %cd% && go run cmd/main.go --id=1 --port=9001 --peers=localhost:9002,localhost:9003"

timeout /t 2 >nul
echo Starting Raft Node 2...
start cmd /k "cd /d %cd% && go run cmd/main.go --id=2 --port=9002 --peers=localhost:9001,localhost:9003"

timeout /t 2 >nul
echo Starting Raft Node 3...
start cmd /k "cd /d %cd% && go run cmd/main.go --id=3 --port=9003 --peers=localhost:9001,localhost:9002"
