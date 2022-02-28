# Compiles .res into .syso
# Run inside res/ directory.

echo "Building id3fit.syso from id3fit.res..."
$GOPATH/src/windres.exe -i id3fit.res -o ../id3fit.syso
echo "Done."