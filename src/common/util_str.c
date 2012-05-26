
#include <stdio.h>
#include "util.h"
#pragma warning(disable:4996)


wchar_t* allocfmtv(const wchar_t *fmt, va_list args)
{
	int len = _vscwprintf(fmt, args); // calculate length, without terminating null
	wchar_t *retbuf = malloc(sizeof(wchar_t) * (len + 1));
	_vsnwprintf(retbuf, len, fmt, args);
	retbuf[len] = L'\0'; // place terminating null
	return retbuf; // must call free() on this
}

wchar_t* allocfmt(const wchar_t *fmt, ...)
{
	wchar_t *retbuf;
	va_list  args;

	va_start(args, fmt);
	retbuf = allocfmtv(fmt, args);
	va_end(args);
  return retbuf; // must call free() on this
}

void appendfmt(wchar_t **pStr, const wchar_t *fmt, ...)
{
	int     oldLen, plusLen;
	va_list args;
	
	va_start(args, fmt);
	oldLen = lstrlen(*pStr);
	plusLen = _vscwprintf(fmt, args); // calculate length, without terminating null
	*pStr = realloc(*pStr, (oldLen + plusLen + 1) * sizeof(wchar_t)); // string is reallocated in-place
	_vsnwprintf(*pStr + oldLen, plusLen, fmt, args);
	(*pStr)[oldLen + plusLen] = L'\0'; // place terminating null
	va_end(args);
}

wchar_t* trim(wchar_t *s)
{
	// LTrim.
	wchar_t *pRun = s;
	while(iswspace(*pRun)) ++pRun;
	if(pRun != s)
		memmove(s, pRun, (lstrlen(pRun) + 1) * sizeof(wchar_t)); // move back

	// RTrim.
	pRun = s + lstrlen(s) - 1; // points to last char of string
	while(iswspace(*pRun)) --pRun;
	*(++pRun) = 0; // truncate string

	return s; // return pointer to same passed string
}

void explodeMultiStr(const wchar_t *multiStr, Strings *pBuf)
{
	// Example multiStr:
	// L"first one\0second one\0third one\0"
	// Will be splitted into an array of pointer to strings.
	// Assumes a well-formed multiStr.

	int numStrings = 0;

	// Count number of null-delimited strings; string end with double null.
	{
		const wchar_t *pRun = multiStr;
		while(*pRun) {
			++numStrings;
			pRun += lstrlen(pRun) + 1;
		}
	}

	// Alloc array of pointers to arrays (strings).
	Strings_realloc(pBuf, numStrings);

	// Alloc and copy each string.
	{
		int i;
		const wchar_t *pRun = multiStr;
		for(i = 0; i < numStrings; ++i) {
			int len = lstrlen(pRun);
			Strings_get(pBuf, i) = malloc(sizeof(wchar_t) * (len + 1));
			memcpy(Strings_get(pBuf, i), pRun, sizeof(wchar_t) * (len + 1));
			pRun += len + 1;
		}
	}
}

void explodeQuotedStr(const wchar_t *quotedStr, Strings *pBuf)
{
	// Example quotedStr:
	// "first one" "second one" "third one"
	// Will be splitted into an array of pointer to strings.
	// Assumes a well-formed quotedStr.

	int numStrings = 0;

	// Count number of quoted strings.
	{
		const wchar_t *pRun = quotedStr;
		for(;;) {
			while(*pRun != L'\"') {
				if(!*pRun) goto no_more_strs;
				++pRun;
			}
			++pRun; // now points to 1st char

			++numStrings;
			while(*pRun != L'\"') ++pRun;
			++pRun; // now points to past closing quote
		}
	}
no_more_strs:

	// Alloc array of pointers to arrays (strings).
	Strings_realloc(pBuf, numStrings);

	// Alloc and copy each string.
	{
		int i;
		const wchar_t *pBase, *pRun = quotedStr;
		for(i = 0; i < numStrings; ++i) {
			int len;

			while(*pRun != L'\"') ++pRun;
			pBase = ++pRun; // now points to 1st char
			while(*pRun != L'\"') ++pRun; // now points to closing quote

			len = pRun - pBase;
			Strings_get(pBuf, i) = malloc(sizeof(wchar_t) * (len + 1));
			memcpy(Strings_get(pBuf, i), pBase, sizeof(wchar_t) * len);
			Strings_get(pBuf, i)[len] = L'\0'; // terminating null

			pBase = ++pRun; // not points to past closing quote
		}
	}
}
