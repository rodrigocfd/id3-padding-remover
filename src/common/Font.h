
#include <Windows.h>


HFONT   Font_create         (const wchar_t *name, int size, BOOL bold, BOOL italic);
HFONT   Font_cloneFromSystem();
void    Font_applyOnChildren(HFONT hFont, HWND hParent);

#define Font_free(hFont) DeleteObject(hFont)
