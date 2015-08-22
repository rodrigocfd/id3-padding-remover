
#include <crtdbg.h>
#include "Glob.h"


Glob Glob_new(const wchar_t *targetDir, const wchar_t *pattern)
{
	Glob obj = { 0 };
	BOOL hasBackslash = targetDir[lstrlen(targetDir) - 1] == '\\';

	obj.pattern = malloc(sizeof(wchar_t) * (
		lstrlen(targetDir) +
		(hasBackslash ? 0 : 1) +
		lstrlen(pattern) + 1 ));
	lstrcpy(obj.pattern, targetDir);
	if(!hasBackslash) lstrcat(obj.pattern, L"\\");
	lstrcat(obj.pattern, pattern); // assembly path + pattern

	return obj;
}

void Glob_free(Glob *pg)
{
	free(pg->pattern);
	if(pg->hFind && pg->hFind != INVALID_HANDLE_VALUE)
		FindClose(pg->hFind);
	SecureZeroMemory(pg, sizeof(Glob));
}

wchar_t* Glob_next(Glob *pg, wchar_t *buf)
{
	// Assumes buf big enough to hold the path strings.

	const wchar_t *pBackslash;
	_ASSERT(pg->pattern);

	if(!pg->hFind) { // first call to method
		if((pg->hFind = FindFirstFile(pg->pattern, &pg->wfd)) == INVALID_HANDLE_VALUE) { // init iteration
			pg->hFind = 0;
			return NULL; // no files found at all
		}
	}
	else { // subsequent calls
		if(!FindNextFile(pg->hFind, &pg->wfd)) {
			FindClose(pg->hFind);
			pg->hFind = 0;
			return NULL; // search finished
		}
	}

	if(pBackslash = wcsrchr(pg->pattern, L'\\')) { // search last backslash on user pattern
		int dirnameLen = (int)(pBackslash - pg->pattern) + 1; // length of directory plus backslash
		lstrcpyn(buf, pg->pattern, dirnameLen + 1); // number of chars includes the terminating null
		lstrcat(buf, pg->wfd.cFileName); // filepath + filename
	}
	else
		lstrcpy(buf, pg->wfd.cFileName); // simply copy
	
	return buf; // same passed buffer
}
