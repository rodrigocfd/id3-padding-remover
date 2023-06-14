# Builds the final application and copies to our local dir.
# Run inside res/ directory.

echo "Building application..."
cd ..
go build -ldflags "-s -w -H=windowsgui"

echo "Deploying to apps folder..."
mv ./id3fit.exe /d/Stuff/apps/_audio\ tools/.

echo "Done."