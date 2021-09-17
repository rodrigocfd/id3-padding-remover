
#include "MainWindow.h"
#include "../res/resource.h"

RUN(MainWindow)

MainWindow::MainWindow()
	: MainDialog{DLG_MAIN, ICO_FROG, 0}
{
}

INT_PTR MainWindow::dialogProc(UINT msg, WPARAM wp, LPARAM lp)
{
	switch (msg) {
	case WM_INITDIALOG: onInitDialog(); return TRUE;
	case WM_COMMAND:
		switch LOWORD(wp) {
		case IDCANCEL: SendMessage(hWnd(), WM_CLOSE, 0, 0); return TRUE;
		}
		break;
	case WM_CLOSE: DestroyWindow(hWnd()); return TRUE;
	case WM_NCDESTROY: PostQuitMessage(0); return TRUE;
	}
	return FALSE;
}

void MainWindow::onInitDialog()
{

}
