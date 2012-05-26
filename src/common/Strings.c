
#include "Strings.h"


Strings Strings_new()
{
	Strings obj = { 0 };
	return obj;
}

void Strings_free(Strings *pStrs)
{
	if(pStrs->ptr) {
		int i;
		for(i = 0; i < pStrs->n; ++i)
			if(pStrs->ptr[i])
				free(pStrs->ptr[i]);
		free(pStrs->ptr);
	}
	SecureZeroMemory(pStrs, sizeof(Strings));
}

void Strings_realloc(Strings *pStrs, int size)
{
	if(!size)
		Strings_free(pStrs);
	else {
		int i;
		for(i = size; i < pStrs->n; ++i)
			free(pStrs->ptr[i]); // when size < n, free excedent strings
		pStrs->ptr = realloc(pStrs->ptr, sizeof(wchar_t*) * size);
		for(i = pStrs->n; i < size; ++i)
			pStrs->ptr[i] = NULL; // when size > n, init new pointers
		pStrs->n = size;
	}
}
