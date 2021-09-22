
#include <core/ListView.h>
#include "MainWindow.h"
#include "../res/resource.h"
using std::optional;
using core::Menu;
using core::ListView;

RUN(MainWindow)

MainWindow::~MainWindow()
{
	this->appMenu.destroy();
}

MainWindow::MainWindow()
	: MainDialog{DLG_MAIN, ICO_FROG, 0},
		iconsList{SIZE{16, 16}},
		appMenu{MNU_FILES}
{
}

INT_PTR MainWindow::dialogProc(UINT msg, WPARAM wp, LPARAM lp)
{
	switch (msg) {
	case WM_INITDIALOG: onInitDialog(); return TRUE;
	case WM_COMMAND:
		switch LOWORD(wp) {
		case IDCANCEL: SendMessage(hWnd(), WM_CLOSE, 0, 0); return TRUE;
		case MNU_FILES_OPEN: onFilesOpen(); return TRUE;
		case MNU_FILES_ABOUT: onFilesAbout(); return TRUE;
		}
		break;
	case WM_NOTIFY:
		if (ListView{hWnd(), LST_FILES, optional{Menu{appMenu.subMenu(0)}}}.onWmNotify(lp)) {
			return TRUE;
		}
		break;
	case WM_CLOSE: DestroyWindow(hWnd()); return TRUE;
	case WM_NCDESTROY: PostQuitMessage(0); return TRUE;
	}
	return FALSE;
}
