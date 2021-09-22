
#include <core/ListView.h>
#include "MainWindow.h"
#include "../res/resource.h"
using std::vector;
using std::wstring;
using core::ListView;

void MainWindow::addFilesToList(const vector<wstring>& mp3s)
{
	ListView lstFiles{hWnd(), LST_FILES};
	lstFiles.setRedraw(false);

	for (const wstring& mp3 : mp3s) {
		lstFiles.items.add(0, {mp3, L"x"});
	}

	lstFiles.setRedraw(true);
}
