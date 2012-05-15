/**
* Program entry-point.
*/

#include <Windows.h>
#include <crtdbg.h>
#include "common/util.h"
#include "common/ListView.h"
#include "../res/resource.h"
#include "MainEvents.h"


static INT_PTR CALLBACK Main_dialogProc(HWND hDlg, UINT msg, WPARAM wp, LPARAM lp)
{
	switch(msg)
	{
	case WM_COMMAND:
		switch(LOWORD(wp))
		{
		case IDCANCEL:        SendMessage(hDlg, WM_CLOSE, 0, 0); return TRUE; // close on ESC
		case MNU_ADDFILES:    Main_onAddFiles(); return TRUE;
		case MNU_DELFROMLIST: ListView_delSelItems(GetDlgItem(hDlg, LST_MAIN)); return TRUE;
		case MNU_SUMMARY:     Main_onSummary(); return TRUE;
		case MNU_REMPADDING:  Main_onRemPadding(); return TRUE;
		}
		break;

	case WM_NOTIFY:
		switch(((NMHDR*)lp)->idFrom)
		{
		case LST_MAIN:
			switch(((NMHDR*)lp)->code)
			{
			case LVN_INSERTITEM:
			case LVN_DELETEALLITEMS: setTextFmt(hDlg, 0, L"ID3 Padding Remover (%d)", ListView_count(GetDlgItem(hDlg, LST_MAIN))); return TRUE; // count items
			case LVN_DELETEITEM:     setTextFmt(hDlg, 0, L"ID3 Padding Remover (%d)", ListView_count(GetDlgItem(hDlg, LST_MAIN)) - 1); return TRUE; // count items			
			case NM_RCLICK:          ListView_popMenu(GetDlgItem(hDlg, LST_MAIN), MEN_MAIN, TRUE); return TRUE; // right-click menu
			case LVN_KEYDOWN:
				switch(((NMLVKEYDOWN*)lp)->wVKey)
				{
				case 'A':       if(hasCtrl()) ListView_selAllItems(GetDlgItem(hDlg, LST_MAIN)); return TRUE; // Ctrl+A
				case VK_DELETE: ListView_delSelItems(GetDlgItem(hDlg, LST_MAIN)); return TRUE; // Del
				case VK_APPS:   popMenu(hDlg, MEN_MAIN, 2, 10, GetDlgItem(hDlg, LST_MAIN)); return TRUE; // context menu
				case VK_F1:     msgBoxFmt(hDlg, MB_ICONINFORMATION, L"About", L"ID3 Padding Remover v1.0.4\nRodrigo César de Freitas Dias."); return TRUE;
				}
				break;
			}
			break;
		}
		break;

	case WM_INITDIALOG:    Main_onInitDialog(hDlg); return TRUE;
	case WM_SIZE:          Main_onSize(wp, lp); return TRUE;
	case WM_DROPFILES:     Main_onDropFiles(wp); return TRUE;
	case WM_INITMENUPOPUP: Main_onInitMenuPopup(wp); return TRUE;
	case WM_CLOSE:         DestroyWindow(hDlg); return TRUE;
	case WM_DESTROY:       PostQuitMessage(0); return TRUE;
	}
	return FALSE;
}

int WINAPI wWinMain(HINSTANCE hInst, HINSTANCE h0, LPWSTR cmdLine, int cmdShow)
{
	int ret = runDialog(hInst, cmdShow, DLG_MAIN, ICO_FROG, 0, Main_dialogProc, 0);
	_ASSERT(!_CrtDumpMemoryLeaks());
	return ret;
}
