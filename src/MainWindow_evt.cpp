
#include <core/ListView.h>
#include "MainWindow.h"
#include "../res/resource.h"
using core::ListView;

void MainWindow::onInitDialog()
{
	iconsList.loadShellIcon({L"mp3"});

	ListView lvFiles{hWnd(), LST_FILES};
	lvFiles.columns.add(L"Files", 0)
		.add(L"Padding", 60)
		.stretch(0);
	lvFiles.setImageList(iconsList, LVSIL_SMALL);

	ListView lvFrames{hWnd(), LST_FRAMES};
	lvFrames.setExtendedStyle(true, LVS_EX_GRIDLINES);
	lvFrames.columns.add(L"Frame", 65)
		.add(L"Value", 0)
		.stretch(1);

}

void MainWindow::onFilesOpen()
{

}

void MainWindow::onFilesAbout()
{

}
