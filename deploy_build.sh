
cd $WORKSPACE || exit

export GOPATH="$WORKSPACE/gopath"
export GOBIN="$GOPATH/bin"

echo "Cleaning previous compiled Go..."
GOCACHE=$WORKSPACE go clean

echo "Installing and building packages..."
## Currently this only builds from the default branch
## When we figure out our approach we can do:
## "GOCACHE=$WORKSPACE go get github.com/tempuslabs/s3s2@GIT_VERSION_NUMBER"
## to pull a specific version
GOCACHE=$WORKSPACE go get github.com/tempuslabs/s3s2

GOOS=linux GOARCH=amd64 GOCACHE=$WORKSPACE go build -o linux/s3s2-linux-amd64 -v &
GOOS=darwin GOARCH=amd64 GOCACHE=$WORKSPACE go build -o darwin/s3s2-darwin-amd64 -v &
GOOS=windows GOARCH=amd64 GOCACHE=$WORKSPACE go build -o windows/s3s2-windows-amd64.exe -v &
