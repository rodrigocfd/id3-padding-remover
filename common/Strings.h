/**
* Automation for a dynamic array of arrays of chars (strings).
*/

#include <Windows.h>

typedef struct {
	int n;
	wchar_t **ptr;
} Strings;


Strings Strings_new       ();
void    Strings_free      (Strings *pStrs);
void    Strings_realloc   (Strings *pStrs, int size);
void    Strings_reallocStr(Strings *pStrs, int i, int lengthWithoutNull);

#define Strings_count(pStrs)     ((pStrs)->n)
#define Strings_get(pStrs, i)    ((pStrs)->ptr[i])
#define Strings_set(pStrs, i, s) ((pStrs)->ptr[i] = _wcsdup(s))
