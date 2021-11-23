
cd $WORKSPACE || exit

export GOPATH="$WORKSPACE/gopath"
export GOBIN="$GOPATH/bin"

echo "Cleaning previous compiled Go..."
GOCACHE=$WORKSPACE go clean

echo "Installing from local Go mod..."
GOCACHE=$WORKSPACE go mod download

echo "Building for operating systems..."
GOOS=linux GOARCH=amd64 GOCACHE=$WORKSPACE go build -o linux/s3s2-linux-amd64 -v
GOOS=darwin GOARCH=amd64 GOCACHE=$WORKSPACE go build -o darwin/s3s2-darwin-amd64 -v
GOOS=windows GOARCH=amd64 GOCACHE=$WORKSPACE go build -o windows/s3s2-windows-amd64.exe -v

GOOS=linux GOARCH=amd64 GOCACHE=$WORKSPACE go build -buildmode=c-shared -o so/s3s2.so sharedobj/sodecrypt.go
