
#include "Worker.h"
#include "../res/resource.h"
#include "WorkerEvents.h"


static INT_PTR CALLBACK WorkerDialog_proc(HWND hDlg, UINT msg, WPARAM wp, LPARAM lp)
{
	switch(msg)
	{
	case WM_INITDIALOG: Worker_onInitDialog(hDlg, lp); return TRUE;
	case WM_FILEDONE:   Worker_onFileDone(); return TRUE;
	case WM_FILEFAILED: Worker_onFileFailed(wp, lp); return TRUE;
	case WM_CLOSE:      EndDialog(hDlg, 0); return TRUE;
	}
	return FALSE;
}

int WorkerDialog_pop(HWND hParent, int numFiles, const wchar_t **pFiles)
{
	WorkerFiles wf = { numFiles, pFiles }; // setup data struct
	return DialogBoxParam(GetModuleHandle(NULL), MAKEINTRESOURCE(DLG_WORKER),
		hParent, WorkerDialog_proc, (LPARAM)&wf); // pass pointer to data struct
}
