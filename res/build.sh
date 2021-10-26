# Builds the final application.

echo "Building application..."
cd ..
go build -ldflags "-s -w -H=windowsgui"

echo "Moving to apps folder..."
mv ./id3fit.exe /d/Stuff/apps/_audio\ tools/.

echo "Done."