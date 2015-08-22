
#include "MainEvents.h"
#include "../common/ListView.h"
#include "../common/Glob.h"
#include "../common/Font.h"
#include "../common/Id3.h"
#include "../common/util.h"
#include "../res/resource.h"
#include "Worker.h"

static HWND hDlg = 0, hList = 0;


static void _Main_calcPaddingSizeAll()
{
	int i, nFiles = ListView_count(hList);
	for(i = 0; i < nFiles; ++i) {
		wchar_t  file[MAX_PATH];
		wchar_t *err;
		Id3      id3 = Id3_new();
		
		ListView_getText(hList, i, 0, file, ARRAYSIZE(file));
		if(!Id3_open(&id3, file, &err)) {
			msgBoxFmt(hDlg, MB_ICONERROR, L"Fail", L"Could not read ID3 padding from:\n%s\n%s", file, err);
			free(err);
			continue;
		}

		ListView_setTextFmt(hList, i, 1, L"%d", Id3_paddingSize(&id3)); // calc the padding of all files
		Id3_free(&id3);
	}
}

void Main_onSize(WPARAM wp, LPARAM lp)
{
	static DWORD padding = 0;

	if(!padding) { // calculate only once
		RECT rcClient = { 0 }, rcList = { 0 };
		GetClientRect(hDlg, &rcClient);
		GetWindowRect(hList, &rcList);
		padding = MAKELONG(rcClient.right - (rcList.right - rcList.left), // right padding
			rcClient.bottom - (rcList.bottom - rcList.top)); // bottom padding
	}

	if(wp != SIZE_MINIMIZED) {
		SetWindowPos(hList, 0, 0, 0, LOWORD(lp) - LOWORD(padding), HIWORD(lp) - HIWORD(padding),
			SWP_NOZORDER | SWP_NOMOVE);
		ListView_fitColumn(hList, 0);
	}
}

void Main_onInitDialog(HWND hDialog)
{
	hDlg = hDialog;
	hList = GetDlgItem(hDlg, LST_MAIN); // keep list handle globally
	Font_applyOnChildren(g_hSysFont, hDlg); // apply global system font

	ListView_fullRowSel(hList);
	ListView_addColumn(hList, L"File", 120);
	ListView_addColumn(hList, L"Padding", 55);
	ListView_fitColumn(hList, 0);
	ListView_pushSysIcon(hList, L"mp3"); // we'll use system's MP3 icon

	Main_onSize(SIZE_MINIMIZED, 0); // first trigger
}

void Main_onDropFiles(WPARAM wp)
{
	HDROP hDrop = (HDROP)wp;
	int i, numFiles = DragQueryFile(hDrop, 0xFFFFFFFF, 0, 0);

	SendMessage(hList, WM_SETREDRAW, (WPARAM)FALSE, 0);
	
	for(i = 0; i < numFiles; ++i) {
		wchar_t path[MAX_PATH];
		DragQueryFile(hDrop, i, path, ARRAYSIZE(path)); // retrieve filepath
		
		if(isDir(path)) { // if folder, add all MP3 files inside of it
			wchar_t subFileBuf[MAX_PATH];
			Glob globMp3 = Glob_new(path, L"*.mp3");

			while(Glob_next(&globMp3, subFileBuf))
				if(endswith(subFileBuf, L".mp3") && !ListView_itemExists(hList, subFileBuf)) // bypass if not MP3, or if already listed
					ListView_addItem(hList, subFileBuf, 0);

			Glob_free(&globMp3);
		}
		else // add single file
			if(endswith(path, L".mp3") && !ListView_itemExists(hList, path)) // bypass if not MP3, or if already listed
				ListView_addItem(hList, path, 0);
	}
	
	_Main_calcPaddingSizeAll();
	SendMessage(hList, WM_SETREDRAW, (WPARAM)TRUE, 0);
	DragFinish(hDrop);
}

void Main_onInitMenuPopup(WPARAM wp)
{
	HMENU hMenu = (HMENU)wp;
	if(GetMenuItemID(hMenu, 0) == MNU_ADDFILES) { // identify by first menu item
		int selCount = ListView_selCount(hList);
		enableMenu(hMenu, MNU_SUMMARY, selCount == 1);
		enableMenu(hMenu, MNU_REMPADDING, selCount > 0);
		enableMenu(hMenu, MNU_DELFROMLIST, selCount > 0);
	}
}

void Main_onAddFiles()
{
	Strings files = Strings_new();

	if(openFiles(hDlg, L"MP3 files (*.mp3)\0*.mp3\0", &files)) {
		int i;
		SendMessage(hList, WM_SETREDRAW, (WPARAM)FALSE, 0);
		
		for(i = 0; i < Strings_count(&files); ++i) {
			if(endswith(Strings_get(&files, i), L".mp3") &&
				!ListView_itemExists(hList, Strings_get(&files, i)) ) // bypass if not MP3, or if already listed
			{
				ListView_addItem(hList, Strings_get(&files, i), 0); // add to list
			}
		}
		
		_Main_calcPaddingSizeAll();
		SendMessage(hList, WM_SETREDRAW, (WPARAM)TRUE, 0);
	}

	Strings_free(&files);
}

void Main_onSummary()
{
	int     i;
	wchar_t file[MAX_PATH], *err = NULL;
	Id3     id3 = Id3_new();
	struct { Id3Frame *ptr; int n; } frames = { 0 };

	// Read selected filename.
	ListView_getText(hList, ListView_getNextSel(hList, -1), 0, file, ARRAYSIZE(file));
	if(!Id3_open(&id3, file, &err)) {
		msgBoxFmt(hDlg, MB_ICONERROR, L"Fail", L"Error parsing file: \n%s\n%s", file, err);
		free(err);
		return;
	}

	// Get the ID3 frames.
	frames.n = Id3_getFrames(&id3, &frames.ptr, &err);
	if(err) {
		msgBoxFmt(hDlg, MB_ICONERROR, L"Fail", L"Could not read the ID3 frames from: %s\n%s", file, err);
		free(err);
		return;
	}

	// Build summary string and pop it out.
	{
		wchar_t *summary = allocfmt(L"Summary (%d):\n", frames.n);
		for(i = 0; i < frames.n; ++i) {
			Id3Frame *theFrame = &frames.ptr[i];
			
			appendfmt(&summary, L"[%s] ", theFrame->name);
			if(theFrame->type == ID3FRAME_TEXT)
				appendfmt(&summary, L"%s\n", Id3Frame_getText(theFrame));
			else if(theFrame->type == ID3FRAME_BINARY)
				appendfmt(&summary, L"%d bytes (%.2f%%%%)\n", Id3Frame_getDataSize(theFrame),
					(float)Id3Frame_getDataSize(theFrame) * 100 / Id3_totalTagSize(&id3) );
		}
		appendfmt(&summary, L"\nTotal: %d bytes.", Id3_totalTagSize(&id3));
		msgBoxFmt(hDlg, MB_ICONINFORMATION, L"Summary", summary);
		free(summary);
	}

	// Cleanup.
	for(i = 0; i < frames.n; ++i)
		Id3Frame_free(&frames.ptr[i]);
	free(frames.ptr);
	Id3_free(&id3);
}

void Main_onRemPadding()
{
	int       i, idx;
	int       nFiles = ListView_selCount(hList);
	wchar_t **pFiles = NULL;
	wchar_t   fileBuf[MAX_PATH];

	if(msgBoxFmt(hDlg, MB_ICONQUESTION | MB_OKCANCEL, L"Remove padding",
		L"Proceed with padding removing of %d file%s?", nFiles, (nFiles > 1 ? L"s" : L"")) != IDOK) return; // prompt user

	pFiles = malloc(sizeof(wchar_t*) * nFiles); // alloc array of strings to be consumed in thread

	i = 0;
	idx = -1;
	while((idx = ListView_getNextSel(hList, idx)) != -1) {
		ListView_getText(hList, idx, 0, fileBuf, ARRAYSIZE(fileBuf));
		pFiles[i++] = _wcsdup(fileBuf); // alloc string
	}
	
	WorkerDialog_pop(hDlg, nFiles, pFiles); // all the process is handled here
	
	// Cleanup.
	for(i = 0; i < nFiles; ++i)
		free(pFiles[i]);
	free(pFiles);

	msgBoxFmt(hDlg, MB_ICONINFORMATION, L"Done",
		nFiles == 1 ? L"%d file has been processed." : L"%d files have been processed.",
		nFiles);
	_Main_calcPaddingSizeAll();
}
