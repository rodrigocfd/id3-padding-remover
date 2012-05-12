/**
* The dialog window responsible for proceed with the padding removing itself,
* using threads, and showing a progress bar.
*/

#include <Windows.h>


int WorkerDialog_pop(HWND hParent, int numFiles, const wchar_t **pFiles);
