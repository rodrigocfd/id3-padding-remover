# Compiles .res into .syso

echo "Building id3fit.syso from compiled.res..."
../../../_gopath/src/windres.exe -i compiled.res -o ../id3fit.syso
echo "Done."