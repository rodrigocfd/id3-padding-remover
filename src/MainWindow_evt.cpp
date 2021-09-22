
#include <core/com.h>
#include <core/ListView.h>
#include "MainWindow.h"
#include <ShObjIdl_core.h> // IFileDialog et al
#include "../res/resource.h"
using std::vector;
using std::wstring;
using core::ComPtr;
using core::ListView;

void MainWindow::onInitDialog()
{
	iconsList.loadShellIcon({L"mp3"});

	ListView lvFiles{hWnd(), LST_FILES};
	lvFiles.columns.add({{L"Files", 0}, {L"Padding", 60}});
	lvFiles.columns.stretch(0);
	lvFiles.setImageList(iconsList, LVSIL_SMALL);

	ListView lvFrames{hWnd(), LST_FRAMES};
	lvFrames.setExtendedStyle(true, LVS_EX_GRIDLINES);
	lvFrames.columns.add({{L"Frame", 65}, {L"Value", 0}});
	lvFrames.columns.stretch(1);
}

void MainWindow::onFilesOpen()
{
	ComPtr<IFileOpenDialog> fod{CLSID_FileOpenDialog};

	FILEOPENDIALOGOPTIONS flags = 0;
	fod->GetOptions(&flags);
	fod->SetOptions(flags | FOS_FORCEFILESYSTEM | FOS_FILEMUSTEXIST | FOS_ALLOWMULTISELECT);

	COMDLG_FILTERSPEC filterSpec[] = {
		{L"MP3 audio files", L"*.mp3"},
		{L"All files", L"*.*"}
	};
	fod->SetFileTypes(ARRAYSIZE(filterSpec), filterSpec);
	fod->SetFileTypeIndex(1);

	if (HRESULT hr = fod->Show(hWnd()); hr == HRESULT_FROM_WIN32(ERROR_CANCELLED)) {
		return;
	}

	ComPtr<IShellItemArray> shItems;
	fod->GetResults(&shItems);

	DWORD numItems = 0;
	shItems->GetCount(&numItems);
		
	vector<wstring> fileNames;
	fileNames.reserve(numItems);
		
	for (DWORD i = 0; i < numItems; ++i) {
		ComPtr<IShellItem> shItem;
		shItems->GetItemAt(i, &shItem);

		LPWSTR pName = nullptr;
		shItem->GetDisplayName(SIGDN_FILESYSPATH, &pName);
		fileNames.emplace_back(pName);
		CoTaskMemFree(pName);
	}

	addFilesToList(fileNames);
}

void MainWindow::onFilesAbout()
{
	TASKDIALOGCONFIG tdc = {0};
	tdc.cbSize = sizeof(TASKDIALOGCONFIG);
	tdc.hwndParent = hWnd();
	tdc.dwFlags = TDF_ALLOW_DIALOG_CANCELLATION;
	tdc.dwCommonButtons = TDCBF_OK_BUTTON;
	tdc.pszMainIcon = TD_INFORMATION_ICON;
	tdc.pszWindowTitle = L"About";
	tdc.pszMainInstruction = L"ID3 Fit";
	tdc.pszContent = L"Rodrigo César de Freitas Dias (C) 2012-2021\n"
		L"Written in C++20 with the Core library.";

	TaskDialogIndirect(&tdc, nullptr, nullptr, nullptr);
}
