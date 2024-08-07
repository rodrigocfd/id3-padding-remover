#include "DlgEdit.h"

INT_PTR DlgEdit::dlgProc(UINT uMsg, WPARAM wp, LPARAM lp)
{
	switch (uMsg) {
		case WM_INITDIALOG: return onInitDialog();
		case WM_CLOSE:      EndDialog(hWnd(), 0); return TRUE; // don't call EndDialog(), so user can't close it
		default:            return FALSE;
	}
}

INT_PTR DlgEdit::onInitDialog()
{

	return TRUE;
}
