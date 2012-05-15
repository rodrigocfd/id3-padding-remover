/**
* Performs a glob operation.
* http://en.wikipedia.org/wiki/Glob_%28programming%29
*/

#include <Windows.h>

typedef struct {
	HANDLE          hFind;
	WIN32_FIND_DATA wfd;
	wchar_t        *pattern;
} Glob;


Glob     Glob_new (const wchar_t *targetDir, const wchar_t *pattern);
void     Glob_free(Glob *pg);
wchar_t* Glob_next(Glob *pg, wchar_t *buf);
