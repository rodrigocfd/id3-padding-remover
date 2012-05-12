
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

static void _releaseParallelBuffers() {
	if(buffer.bUsed) {
		free(buffer.pFiles);
		free(buffer.pFolder);
		buffer.pFiles = buffer.pFolder = NULL;
		buffer.bUsed = FALSE;
	}
}

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
					_releaseParallelBuffers(); // original buffer is larger enough, we don't need the parallel buffers
			}
			break;
		}
		break;
	}
	return 0;
}

int openFiles(HWND hWnd, const wchar_t *filter, wchar_t ***pBuf)
{
	OPENFILENAME ofn = { 0 };
	BOOL         retCode;
	wchar_t      multiBuf[256] = { 0 }; // will receive the multi-string
	int          numFiles = 0;

	ofn.lStructSize = sizeof(ofn);
	ofn.hwndOwner   = hWnd;
	ofn.lpstrFilter = filter;
	ofn.lpstrFile   = multiBuf;
	ofn.nMaxFile    = ARRAYSIZE(multiBuf);
	ofn.lpfnHook    = _openFilesHookProc;
	ofn.Flags       = OFN_FILEMUSTEXIST | OFN_ALLOWMULTISELECT |
		OFN_EXPLORER | OFN_ENABLESIZING | OFN_ENABLEHOOK;

	*pBuf = NULL;
	retCode = GetOpenFileName(&ofn);

	if( (retCode && buffer.bUsed) || (!retCode && CommDlgExtendedError() == FNERR_BUFFERTOOSMALL) )
	{
		int i;
		wchar_t **pParsed = NULL;

		numFiles = quotedstr2array(buffer.pFiles, &pParsed); // break quoted-string into string array
		_ASSERT(numFiles);
		*pBuf = malloc(sizeof(wchar_t*) * numFiles); // alloc return buffer

		for(i = 0; i < numFiles; ++i) {
			(*pBuf)[i] = malloc(sizeof(wchar_t) *
				(lstrlen(buffer.pFolder) + lstrlen(pParsed[i]) + 2)); // room for backslash and null
			lstrcpy((*pBuf)[i], buffer.pFolder);
			lstrcat((*pBuf)[i], L"\\");
			lstrcat((*pBuf)[i], pParsed[i]); // concat folder + file
		}

		for(i = 0; i < numFiles; ++i) free(pParsed[i]); // cleanup
		free(pParsed);
	}
	else if(retCode) // call OK with regular stack buffer
	{
		int i;
		wchar_t *pBasePath;
		struct { int num; wchar_t **ptr; } strs = { 0 };
		
		strs.num = multistr2array(multiBuf, &strs.ptr); // break multi-string into string array
		_ASSERT(strs.num);

		if(strs.num == 1) // if user selected only 1 file, the string is the full path, and that's all
		{
			numFiles = 1;
			*pBuf = malloc(sizeof(wchar_t*) * 1); // alloc return buffer; array of 1 string
			(*pBuf)[0] = malloc(sizeof(wchar_t) * (lstrlen(strs.ptr[0]) + 1)); // alloc unique string
			lstrcpy((*pBuf)[0], strs.ptr[0]);
		}
		else // user selected 2 or more files
		{
			pBasePath = strs.ptr[0]; // 1st string is the base path; others are the filenames
			numFiles = strs.num - 1;
			*pBuf = malloc(sizeof(wchar_t*) * numFiles); // alloc return buffer

			for(i = 0; i < numFiles; ++i) {
				(*pBuf)[i] = malloc(sizeof(wchar_t) *
					(lstrlen(pBasePath) + lstrlen(strs.ptr[i + 1]) + 2)); // room for backslash and null
				lstrcpy((*pBuf)[i], pBasePath);
				lstrcat((*pBuf)[i], L"\\");
				lstrcat((*pBuf)[i], strs.ptr[i + 1]); // concat folder + file
			}
		}
		
		for(i = 0; i < strs.num; ++i) free(strs.ptr[i]); // cleanup
		free(strs.ptr);
	}

	_releaseParallelBuffers(); // eventual cleanup
	return numFiles; // user must free this array of arrays, if files have been returned
}
