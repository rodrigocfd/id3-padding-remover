# Compiles .res into .syso

echo "Building id3fit.syso from compiled.res..."
$GOPATH/src/windres.exe -i compiled.res -o ../id3fit.syso
echo "Done."