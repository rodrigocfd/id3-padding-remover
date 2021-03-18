# Builds the final application.

echo "Building application..."
cd ..
go build -ldflags "-s -w -H=windowsgui"
echo "Done."