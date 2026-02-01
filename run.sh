# kill anything running on 8080
sudo lsof -t -i tcp:8080 | xargs sudo kill -9
echo "Killed any process running on port 8080"
# run the server
go run main.go
