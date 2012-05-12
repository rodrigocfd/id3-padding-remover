
#include "WorkerEvents.h"
#include <CommCtrl.h>
#include "../res/resource.h"
#include "common/util.h"
#include "common/Id3.h"
#include "common/Font.h"
#include "common/Thread.h"

static HWND hDlg = 0;
static WorkerFiles *pWf = 0;


static void _Worker_processFile(void *arg)
{
	int      idx = (int)arg;
	Id3      id3 = Id3_new();
	wchar_t *err = NULL;

	Sleep(50 * idx); // alleviate the heavy disk concurrency (and make it look pretty)
	Id3_open(&id3, pWf->ptr[idx], &err);
	if(err) {
		SendMessage(hDlg, WM_FILEFAILED, 0, (LPARAM)err); // in case of error, tell parent
		free(err);
	}
	else {
		Id3_removePadding(&id3);
		Id3_free(&id3);
	}

	PostMessage(hDlg, WM_FILEDONE, (WPARAM)idx, 0); // tell parent we've passed another one
}

void Worker_onInitDialog(HWND hDialog, LPARAM lp)
{
	int i;

	hDlg = hDialog;
	pWf = (WorkerFiles*)lp; // retrieve and store pointer to structure
	centerOnParent(hDlg);
	Font_applyOnChildren(g_hSysFont, hDlg); // apply global system font
	enableMenu(GetSystemMenu(hDlg, FALSE), SC_CLOSE, FALSE); // disable X button
	setTextFmt(hDlg, LBL_STATUS, L"Processing file 0 of %d...", pWf->n);
	
	SendDlgItemMessage(hDlg, PRO_STATUS, PBM_SETRANGE, 0, MAKELPARAM(0, pWf->n - 1)); // setup progress bar
	SendDlgItemMessage(hDlg, PRO_STATUS, PBM_SETPOS, (WPARAM)0, 0);

	for(i = 0; i < pWf->n; ++i)
		Thread_RunAsync(_Worker_processFile, (void*)i); // pass just the index to the thread callback
}

void Worker_onFileDone()
{
	static int filesDone = 0;
	SendDlgItemMessage(hDlg, PRO_STATUS, PBM_SETPOS, (WPARAM)filesDone, 0); // setup progress bar
	setTextFmt(hDlg, LBL_STATUS, L"Processing file %d of %d...", ++filesDone, pWf->n);
	if(filesDone == pWf->n) {
		filesDone = 0; // reset for further use
		pWf = NULL;
		SendMessage(hDlg, WM_CLOSE, 0, 0); // close window after last file
	}
}

void Worker_onFileFailed(WPARAM wp, LPARAM lp)
{
	msgBoxFmt(hDlg, MB_ICONERROR, L"Fail",
		L"The file could not be processed:\n%s\n%s",
		pWf->ptr[(int)wp], (const wchar_t*)lp);
}
