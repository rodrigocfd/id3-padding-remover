# Compiles .res into .syso

echo "Building id3fit.syso from id3-fit.res..."
$GOPATH/src/windres.exe -i id3-fit.res -o ../id3-fit.syso
echo "Done."