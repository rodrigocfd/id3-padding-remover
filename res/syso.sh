# Compiles .res into .syso

PROJ=id3-fit

echo "Building $PROJ.syso from $PROJ.res..."
$GOPATH/../syso/windres.exe -i $PROJ.res -o $PROJ.syso
mv $PROJ.syso ..
echo "Done."
