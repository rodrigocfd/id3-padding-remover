
#include <Windows.h>

typedef struct Glob_ {
	HANDLE          hFind;
	WIN32_FIND_DATA wfd;
	wchar_t        *pattern;
} Glob;


Glob     Glob_new (const wchar_t *targetDir, const wchar_t *pattern);
void     Glob_free(Glob *pg);
wchar_t* Glob_next(Glob *pg, wchar_t *buf);
