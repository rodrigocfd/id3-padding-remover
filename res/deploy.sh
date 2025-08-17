# Builds the final application and copies to our local dir.
# Run inside res/ directory.

echo "Building application..."
cd ..
go build -ldflags "-s -w -H=windowsgui"

EXE=id3-fit_GO.exe
echo "Deploying to apps folder..."
mv ./id3fit.exe /d/Stuff/apps/_audio\ tools/$EXE

/c/apps-portable/Go/gsa /d/Stuff/apps/_audio\ tools/$EXE
