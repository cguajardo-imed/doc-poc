# Set the CGO_ENABLED environment variable to 1
$env:CGO_ENABLED = 1

# Execute the go build command
go build -o ./tmp/main.exe .