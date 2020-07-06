Generate syso
-------------

1) Compile the *.rc file into *.res with Visual C++ Developer Power Shell for VS 2019:
rc /r resources.rc

2) On MinGW prompt, convert *.res into *.syso:
../../windres -i resources.res -o resources.syso

3) Place *.syso in the same directory of main.go and build the *.exe:
go build -ldflags "-s -w -H=windowsgui"

---
http://bluedesk.blogspot.com/2020/06/embedding-rc-files-into-go-win32.html