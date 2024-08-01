#include "DlgMain.h"
#include "../res/resource.h"

int APIENTRY wWinMain(_In_ HINSTANCE hInst, _In_opt_ HINSTANCE, _In_ LPWSTR, _In_ int cmdShow)
{
	lib::ComOle oleLib;
	DlgMain d;
	return lib::runMain(d, hInst, DLG_MAIN, cmdShow, ICO_FOULBACHELOR);
}

INT_PTR DlgMain::dlgProc(UINT uMsg, WPARAM wp, LPARAM lp)
{
	switch (uMsg) {
		case WM_INITDIALOG: return onInitDialog();
		case WM_CLOSE:      DestroyWindow(hWnd()); return TRUE;
		case WM_NCDESTROY:  PostQuitMessage(0); return TRUE;
		default:            return FALSE;
	}
}

INT_PTR DlgMain::onInitDialog()
{
	return TRUE;
}
