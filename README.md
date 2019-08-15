# ID3 Padding Remover

When you attach ID3v2 tags to an MP3 file, usually the program writes some padding bytes as well, so that future changes on the tag would be faster. But I usually tag my files only once (properly, right after I grab them), so I won't ever touch them again. Then these padding bytes are just a waste of hard disk space.

I use Mp3tag to write my tags, and it always writes padding bytes. So I decided to write a program to scan the MP3 files and remove the padding bytes, shrinking down the file size to its optimal. It currently supports only ID3v2.3.0 tags (the one I use).

This program is pure C, and it uses nothing but Win32. It works well on Wine.
