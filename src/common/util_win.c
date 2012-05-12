
#include "util.h"
#include "Font.h"
#include <CommCtrl.h>
#pragma comment(linker, \
  "\"/manifestdependency:type='Win32' "\
  "name='Microsoft.Windows.Common-Controls' "\
  "version='6.0.0.0' "\
  "processorArchitecture='*' "\
  "publicKeyToken='6595b64144ccf1df' "\
  "language='*'\"")
#pragma comment(lib, "ComCtl32.lib")


HFONT g_hSysFont = 0; // global system font definition


void debugfmt(const wchar_t *fmt, ...)
{
	wchar_t *buf;
	va_list  args;

	va_start(args, fmt);
	buf = allocfmtv(fmt, args);
	va_end(args);
	OutputDebugString(buf);
	free(buf);
}

void setTextFmt(HWND hWnd, int id, const wchar_t *fmt, ...)
{
	wchar_t *buf;
	va_list  args;

	va_start(args, fmt);
	buf = allocfmtv(fmt, args);
	va_end(args);

	if(id) SetDlgItemText(hWnd, id, buf);
	else SetWindowText(hWnd, buf); // if ID is zero, just set to the window itself

	free(buf);
}

static HHOOK _hHookMsgBox = 0;
static LRESULT CALLBACK _msgBoxHookProc(int code, WPARAM wp, LPARAM lp)
{
	// http://www.codeguru.com/cpp/w-p/win32/messagebox/print.php/c4541
	if(code == HCBT_ACTIVATE)
	{
		HWND hMsgbox = (HWND)wp;
		HWND hParent = GetForegroundWindow();
		RECT rcMsgbox, rcParent;

		if(hMsgbox && hParent && GetWindowRect(hMsgbox, &rcMsgbox) && GetWindowRect(hParent, &rcParent))
		{
			RECT  rcScreen = { 0 };
			POINT pos = { 0 };
			
			SystemParametersInfo(SPI_GETWORKAREA, 0, (PVOID)&rcScreen, 0); // size of desktop

			// Adjusted x,y coordinates to message box window.
			pos.x = rcParent.left + (rcParent.right - rcParent.left) / 2 - (rcMsgbox.right - rcMsgbox.left) / 2;
			pos.y =	rcParent.top + (rcParent.bottom - rcParent.top) / 2 - (rcMsgbox.bottom - rcMsgbox.top) / 2;

			// Screen out-of-bounds corrections.
			if(pos.x < 0)
				pos.x = 0;
			else if(pos.x + (rcMsgbox.right - rcMsgbox.left) > rcScreen.right)
				pos.x = rcScreen.right - (rcMsgbox.right - rcMsgbox.left);
			if(pos.y < 0)
				pos.y = 0;
			else if(pos.y + (rcMsgbox.bottom - rcMsgbox.top) > rcScreen.bottom)
				pos.y = rcScreen.bottom - (rcMsgbox.bottom - rcMsgbox.top);

			MoveWindow(hMsgbox, pos.x, pos.y,
				rcMsgbox.right - rcMsgbox.left, rcMsgbox.bottom - rcMsgbox.top,
				FALSE);
		}
		UnhookWindowsHookEx(_hHookMsgBox); // release hook
	}
	return CallNextHookEx(0, code, wp, lp);
}

int msgBox(HWND hParent, UINT uType, const wchar_t *caption, const wchar_t *msg)
{
	// The hook is set to center the message box window on parent.
	_hHookMsgBox = SetWindowsHookEx(WH_CBT, _msgBoxHookProc, 0, GetCurrentThreadId());
	return MessageBox(hParent, msg, caption, uType);
}

int msgBoxFmt(HWND hParent, UINT uType, const wchar_t *caption, const wchar_t *msg, ...)
{
	wchar_t *buf;
	va_list  args;
	int      ret;

	va_start(args, msg);
	buf = allocfmtv(msg, args);
	va_end(args);
	ret = msgBox(hParent, uType, caption, buf);
	free(buf);
	return ret;
}

int runDialog(HINSTANCE hInst, int cmdShow, int dialogId, int iconId, int accelTableId, DLGPROC dialogProc, LPARAM lp)
{
	HWND   hDlg = NULL;
	HACCEL hAccel = NULL;
	MSG    msg = { 0 };
	BOOL   ret = FALSE;
	HICON  hIcon16 = NULL, hIcon32 = NULL;

	InitCommonControls();
	g_hSysFont = Font_cloneFromSystem(); // retrieve global system font
	hDlg = CreateDialogParam(hInst, MAKEINTRESOURCE(dialogId), 0, dialogProc, lp);

	if(iconId) { // put icon on dialog system menu, if any (probably yes)
		hIcon16 = (HICON)LoadImage(hInst, MAKEINTRESOURCE(iconId), IMAGE_ICON, 16, 16, LR_DEFAULTCOLOR);
		hIcon32 = (HICON)LoadImage(hInst, MAKEINTRESOURCE(iconId), IMAGE_ICON, 32, 32, LR_DEFAULTCOLOR);
		SendMessage(hDlg, WM_SETICON, ICON_SMALL, (LPARAM)hIcon16);
		SendMessage(hDlg, WM_SETICON, ICON_BIG, (LPARAM)hIcon32);
	}

	ShowWindow(hDlg, cmdShow);

	if(accelTableId) // load accelerators table, if any
		hAccel = LoadAccelerators(hInst, MAKEINTRESOURCE(accelTableId));
	
	while((ret = GetMessage(&msg, 0, 0, 0)) != 0) {
		if(ret == -1) return -1; // failure
		if(!(hAccel && TranslateAccelerator(hDlg, hAccel, &msg)) && !IsDialogMessage(hDlg, &msg)) {
			TranslateMessage(&msg);
			DispatchMessage(&msg);
		}
	}

	if(hIcon16) DestroyIcon(hIcon16); // release the dialog icons
	if(hIcon32) DestroyIcon(hIcon32);
	Font_free(g_hSysFont); // release global system font
	return (int)msg.wParam; // this can be the return value of the program
}

void centerOnParent(HWND hDlg)
{
	// This function centers a child popup on its parent.
	// It works better when called within WM_INITDIALOG.
	HWND hParent = GetParent(hDlg);
	RECT rcParent, rcDlg;

	GetWindowRect(hParent, &rcParent);
	GetWindowRect(hDlg, &rcDlg);
	SetWindowPos(hDlg, 0,
		(rcParent.right - rcParent.left) / 2 + rcParent.left - (rcDlg.right - rcDlg.left) / 2,
		(rcParent.bottom - rcParent.top) / 2 + rcParent.top - (rcDlg.bottom - rcDlg.top) / 2,
		0, 0, SWP_NOZORDER | SWP_NOSIZE);
}

void popMenu(HWND hDlg, int popupMenuId, int x, int y, HWND hWndCoordsRelativeTo)
{
	HMENU hMenu = LoadMenu(GetModuleHandle(NULL), MAKEINTRESOURCE(popupMenuId));
	POINT ptDlg = { x, y }; // receives coordinates relative to hDlg
	ClientToScreen(hWndCoordsRelativeTo ? hWndCoordsRelativeTo : hDlg, &ptDlg); // to screen coordinates
	SetForegroundWindow(hDlg);
	TrackPopupMenu(GetSubMenu(hMenu, 0), 0, ptDlg.x, ptDlg.y, 0, hDlg, NULL); // owned by dialog, so messages go to it
	PostMessage(hDlg, WM_NULL, 0, 0); // http://msdn.microsoft.com/en-us/library/ms648002%28VS.85%29.aspx
	DestroyMenu(hMenu);
}

HICON explorerIcon(const wchar_t *fileExtension)
{
	wchar_t    extens[10];
	SHFILEINFO shfi = { 0 };

	lstrcpy(extens, L"*.");
	lstrcat(extens, fileExtension); // prefix extension; user pass just the 3 letters (or 4, or whathever)

	SHGetFileInfo(extens, FILE_ATTRIBUTE_NORMAL, &shfi, sizeof(shfi),
		SHGFI_TYPENAME | SHGFI_USEFILEATTRIBUTES);
	SHGetFileInfo(extens, FILE_ATTRIBUTE_NORMAL, &shfi, sizeof(shfi),
		SHGFI_ICON | SHGFI_SMALLICON | SHGFI_SYSICONINDEX | SHGFI_USEFILEATTRIBUTES);

	return shfi.hIcon; // user must call DestroyIcon() on this
}
