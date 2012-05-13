
#include "util.h"
#include "ListView.h"


void ListView_addColumn(HWND hList, const wchar_t *caption, int cx)
{
	LVCOLUMN lvc = { 0 };
	lvc.mask = LVCF_TEXT | LVCF_MINWIDTH | LVCF_WIDTH;
	lvc.pszText = (wchar_t*)caption;
	lvc.cx = lvc.cxMin = cx;
	ListView_InsertColumn(hList, 0xFFFF, &lvc);
}

void ListView_fitColumn(HWND hList, int iCol)
{
	int numCols = Header_GetItemCount(ListView_GetHeader(hList));
	int i, cxUsed = 0;
	RECT rc;

	for(i = 0; i < numCols; ++i) {
		if(i != iCol) {
			LVCOLUMN lvc = { 0 };
			lvc.mask = LVCF_WIDTH;
			ListView_GetColumn(hList, i, &lvc); // retrieve cx of each column, except stretchee
			cxUsed += lvc.cx; // sum up
		}
	}
	GetClientRect(hList, &rc); // listview client area
	ListView_SetColumnWidth(hList, iCol,
		rc.right - GetSystemMetrics(SM_CXVSCROLL) - cxUsed); // fit the rest of available space
}

int ListView_addItem(HWND hList, const wchar_t *caption, int iconIdx)
{
	LVITEM lvi = { 0 };

	lvi.iItem   = 0x0FFFFFFF;
	lvi.mask    = LVIF_TEXT | (iconIdx == -1 ? 0 : LVIF_IMAGE);
	lvi.pszText = (wchar_t*)caption;
	lvi.iImage  = iconIdx;
	
	return ListView_InsertItem(hList, &lvi); // return index of newly inserted item
}

static HIMAGELIST _ListView_proceedImageList(HWND hList)
{
	// Imagelist is destroyed automatically:
	// http://www.catch22.net/tuts/sysimgq
	// http://www.autohotkey.com/docs/commands/ListView.htm

	HIMAGELIST hImg = ListView_GetImageList(hList, LVSIL_SMALL); // current imagelist
	if(!hImg) {
		hImg = ImageList_Create(16, 16, ILC_COLOR32, 1, 1); // create a 16x16 imagelist
		if(!hImg) return NULL; // imagelist creation failure!
		ListView_SetImageList(hList, hImg, LVSIL_SMALL); // associate imagelist to listview control
	}
	return hImg; // return handle to current imagelist
}

int ListView_pushIcon(HWND hList, int iconId)
{
	HIMAGELIST hImg = _ListView_proceedImageList(hList);
	HICON icon = (HICON)LoadImage(GetModuleHandle(NULL),
		MAKEINTRESOURCE(iconId), IMAGE_ICON, 16, 16, LR_DEFAULTCOLOR);
	int idx = ImageList_AddIcon(hImg, icon);
	DestroyIcon(icon);
	return idx; // return the index of the new icon
}

int ListView_pushSysIcon(HWND hList, const wchar_t *fileExtension)
{
	HIMAGELIST hImg = _ListView_proceedImageList(hList);
	HICON expicon = explorerIcon(fileExtension);
	int idx = ImageList_AddIcon(hImg, expicon);
	DestroyIcon(expicon);
	return idx; // return the index of the new icon
}

BOOL ListView_itemExists(HWND hList, const wchar_t *caption)
{
	LVFINDINFO lfi = { 0 };
	lfi.flags = LVFI_STRING; // search is case-insensitive
	lfi.psz = caption;
	return ListView_FindItem(hList, -1, &lfi) != -1;
}

void ListView_delSelItems(HWND hList)
{
	int i = -1;
	SendMessage(hList, WM_SETREDRAW, (WPARAM)FALSE, 0);
	while((i = ListView_GetNextItem(hList, -1, LVNI_SELECTED)) != -1)
		ListView_DeleteItem(hList, i);
	SendMessage(hList, WM_SETREDRAW, (WPARAM)TRUE, 0);
}

void ListView_setTextFmt(HWND hList, int i, int col, const wchar_t *fmt, ...)
{
	wchar_t *buf;
	va_list  args;

	va_start(args, fmt);
	buf = allocfmtv(fmt, args);
	va_end(args);
	ListView_SetItemText(hList, i, col, (wchar_t*)buf);
	free(buf);
}

int ListView_popMenu(HWND hList, int popupMenuId, BOOL popEvenWithoutItem)
{
	LVHITTESTINFO lvhti = { 0 };
	int iItem;
	
	GetCursorPos(&lvhti.pt);
	ScreenToClient(hList, &lvhti.pt); // current cursor position
	iItem = ListView_HitTest(hList, &lvhti); // item below cursor, if any

	// The popup menu is created with hDlg as parent, so the menu messages go to it.
	// The lvhti coordinates are relative to hList, and will be mapped into screen-relative.
	if(popupMenuId && !(iItem == -1 && !popEvenWithoutItem))
		popMenu(GetParent(hList), popupMenuId, lvhti.pt.x, lvhti.pt.y, hList);
	
	return iItem; // returns item below cursor, -1 if none, call none() to check
}
