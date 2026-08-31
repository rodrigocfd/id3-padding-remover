# Compiles .res into .syso
# Run inside res/ directory.

PROJ=id3-fit

echo "Building $PROJ.syso from $PROJ.res..."
$GOPATH/../syso/windres.exe -i $PROJ.res -o $PROJ.syso
mv $PROJ.syso ..

echo "Done."
read -n 1 -s -r -p "Press any key to exit..."
