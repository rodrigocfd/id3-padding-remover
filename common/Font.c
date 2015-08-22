
#include "Font.h"


HFONT Font_create(const wchar_t *name, int size, BOOL bold, BOOL italic)
{
	LOGFONT lf = { 0 };
	lstrcpy(lf.lfFaceName, name);
	lf.lfHeight = -(size + 3);
	lf.lfWeight = bold ? FW_BOLD : FW_DONTCARE;
	lf.lfItalic = (BYTE)italic;
	return CreateFontIndirect(&lf);
}

HFONT Font_cloneFromSystem()
{
	OSVERSIONINFO ovi = { 0 };
	NONCLIENTMETRICS ncm = { 0 };

	ovi.dwOSVersionInfoSize = sizeof(ovi);
	GetVersionEx(&ovi);
	
	ncm.cbSize = sizeof(ncm);
	if(ovi.dwMajorVersion < 6) // below Vista
		ncm.cbSize -= sizeof(ncm.iBorderWidth);

	SystemParametersInfo(SPI_GETNONCLIENTMETRICS, ncm.cbSize, &ncm, 0);
	return CreateFontIndirect(&ncm.lfMenuFont);
}

static BOOL CALLBACK _Font_applyOnSingleChild(HWND hWnd, LPARAM lp)
{
	SendMessage(hWnd, WM_SETFONT, (WPARAM)(HFONT)lp, MAKELPARAM(FALSE, 0)); // will run on each child
	return TRUE;
}

void Font_applyOnChildren(HFONT hFont, HWND hParent)
{
	EnumChildWindows(hParent, _Font_applyOnSingleChild, (LPARAM)hFont); // propagate to children
}
