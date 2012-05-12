
#include <Windows.h>
#define WM_FILEDONE   WM_USER + 1
#define WM_FILEFAILED WM_USER + 2


typedef struct WorkerFiles_ {
	int n;
	const wchar_t **ptr;
} WorkerFiles;

void Worker_onInitDialog(HWND hDialog, LPARAM lp);
void Worker_onFileDone();
void Worker_onFileFailed(WPARAM wp, LPARAM lp);
