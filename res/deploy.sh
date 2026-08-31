# Builds the final program.
# Run inside res/ directory.

echo "Building..."
cd ..
go build -trimpath -ldflags "-s -w -H=windowsgui"

/c/app-portable/Go/gsa.exe id3fit.exe

read -n 1 -s -r -p "Press any key to exit..."
