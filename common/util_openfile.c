
#include <crtdbg.h>
#include "util.h"


BOOL openFile(HWND hWnd, const wchar_t *filter, wchar_t *buf, int szBuf)
{
	OPENFILENAME ofn = { 0 };

	*buf = 0;
	ofn.lStructSize = sizeof(ofn);
	ofn.hwndOwner   = hWnd;
	ofn.lpstrFilter = filter;
	ofn.lpstrFile   = buf;
	ofn.nMaxFile    = szBuf;
	ofn.Flags       = OFN_EXPLORER | OFN_ENABLESIZING | OFN_FILEMUSTEXIST;// | OFN_HIDEREADONLY;

	return GetOpenFileName(&ofn) != 0;
}


struct { // parallel buffers
	wchar_t *pFiles;
	wchar_t *pFolder;
	BOOL     bUsed;
} buffer = { 0 };

static UINT_PTR CALLBACK _openFilesHookProc(HWND hWnd, UINT msg, WPARAM wp, LPARAM lp)
{
	// http://www.codeproject.com/Articles/3235/Multiple-Selection-in-a-File-Dialog
	static OPENFILENAME *ofn = NULL;

	switch(msg)
	{
	case WM_INITDIALOG:
		ofn = (OPENFILENAME*)lp; // keep pointer to origin struct
		break;
	case WM_NOTIFY:
		switch(((OFNOTIFY*)lp)->hdr.code)
		{
		//case CDN_FILEOK:
		case CDN_SELCHANGE:
			{
				HWND hWndSel = GetParent(hWnd);
				wchar_t dummyBuf;
				UINT szFiles = CommDlg_OpenSave_GetSpec(hWndSel, &dummyBuf, 1); // get required sizes
				UINT szFolder = CommDlg_OpenSave_GetFolderPath(hWndSel, &dummyBuf, 1);

				if(szFiles + szFolder > ofn->nMaxFile) { // we're larger? let's use parallel buffers
					buffer.bUsed = TRUE;
					buffer.pFiles = realloc(buffer.pFiles, sizeof(wchar_t) * (szFiles + 1));
					CommDlg_OpenSave_GetSpec(hWndSel, buffer.pFiles, szFiles); // grab string

					buffer.pFolder = realloc(buffer.pFolder, sizeof(wchar_t) * (szFolder + 1));
					CommDlg_OpenSave_GetFolderPath(hWndSel, buffer.pFolder, szFolder);
				}
				else
					buffer.bUsed = FALSE; // original buffer is larger enough, we don't need the parallel buffers
			}
			break;
		}
		break;
	}
	return 0;
}

BOOL openFiles(HWND hWnd, const wchar_t *filter, Strings *pBuf)
{
	OPENFILENAME ofn = { 0 };
	BOOL         retCode = FALSE, retCodeNow = FALSE;
	wchar_t      multiBuf[256] = { 0 }; // will receive the multi-string

	ofn.lStructSize = sizeof(ofn);
	ofn.hwndOwner   = hWnd;
	ofn.lpstrFilter = filter;
	ofn.lpstrFile   = multiBuf;
	ofn.nMaxFile    = ARRAYSIZE(multiBuf);
	ofn.lpfnHook    = _openFilesHookProc;
	ofn.Flags       = OFN_FILEMUSTEXIST | OFN_ALLOWMULTISELECT |
		OFN_EXPLORER | OFN_ENABLESIZING | OFN_ENABLEHOOK;

	Strings_realloc(pBuf, 0);
	retCode = GetOpenFileName(&ofn);

	if( (retCode && buffer.bUsed) || (!retCode && CommDlgExtendedError() == FNERR_BUFFERTOOSMALL) )
	{
		int     i;
		Strings parsedFiles = Strings_new();

		explodeQuotedStr(buffer.pFiles, &parsedFiles);
		_ASSERT(Strings_count(&parsedFiles));
		Strings_realloc(pBuf, Strings_count(&parsedFiles)); // alloc return buffer

		for(i = 0; i < Strings_count(&parsedFiles); ++i) {
			Strings_reallocStr(pBuf, i,
				lstrlen(buffer.pFolder) + lstrlen(Strings_get(&parsedFiles, i)) + 1); // room for backslash
			lstrcpy(Strings_get(pBuf, i), buffer.pFolder);
			lstrcat(Strings_get(pBuf, i), L"\\");
			lstrcat(Strings_get(pBuf, i), Strings_get(&parsedFiles, i)); // concat folder + file
		}

		Strings_free(&parsedFiles);
		retCodeNow = TRUE; // okay
	}
	else if(retCode) // call OK with regular stack buffer
	{
		int      i;
		Strings  strs = Strings_new();
		wchar_t *pBasePath = NULL;
		
		explodeMultiStr(multiBuf, &strs);
		_ASSERT(Strings_count(&strs));

		if(Strings_count(&strs) == 1) // if user selected only 1 file, the string is the full path, and that's all
		{
			Strings_realloc(pBuf, 1); // alloc return buffer; array of 1 string
			Strings_set(pBuf, 0, Strings_get(&strs, 0));
		}
		else // user selected 2 or more files
		{
			pBasePath = Strings_get(&strs, 0); // 1st string is the base path; others are the filenames
			Strings_realloc(pBuf, Strings_count(&strs) - 1); // alloc return buffer

			for(i = 0; i < Strings_count(&strs) - 1; ++i) {
				Strings_reallocStr(pBuf, i,
					lstrlen(pBasePath) + lstrlen(Strings_get(&strs, i + 1)) + 1); // room for backslash
				lstrcpy(Strings_get(pBuf, i), pBasePath);
				lstrcat(Strings_get(pBuf, i), L"\\");
				lstrcat(Strings_get(pBuf, i), Strings_get(&strs, i + 1)); // concat folder + file
			}
		}
		
		Strings_free(&strs);
		retCodeNow = TRUE; // okay
	}

	if(buffer.pFiles) free(buffer.pFiles); // eventual cleanup
	if(buffer.pFolder) free(buffer.pFolder);
	buffer.pFiles = buffer.pFolder = NULL;
	buffer.bUsed = FALSE;

	return retCodeNow;
}
