
#include <Windows.h>


/* String-related stuff. */
#define same(a, b) (!lstrcmp(a, b))
#define endswith(s, end) same((s) + lstrlen(s) - lstrlen(end), end)
#define within(x, a, b) ((x) >= (a) && (x) <= (b))
#define between(x, a, b) ((x) > (a) && (x) < (b))

wchar_t* allocfmtv(const wchar_t *fmt, va_list args);
wchar_t* allocfmt(const wchar_t *fmt, ...);
void     appendfmt(wchar_t **pStr, const wchar_t *fmt, ...);
wchar_t* trim(wchar_t *s);
int multistr2array(const wchar_t *multiStr, wchar_t ***pBuf);
int quotedstr2array(const wchar_t *quotedStr, wchar_t ***pBuf);


/* Win32 shorthand routines. */
#define isDir(path) ((GetFileAttributes(path) & FILE_ATTRIBUTE_DIRECTORY) != 0)
#define enableMenu(hmenu, id, enable) EnableMenuItem(hmenu, id, MF_BYCOMMAND | ((enable) ? MF_ENABLED : MF_GRAYED))
#define hasCtrl() ((GetAsyncKeyState(VK_CONTROL) & 0x8000) != 0)

void  debugfmt(const wchar_t *fmt, ...);
void  setTextFmt(HWND hWnd, int id, const wchar_t *fmt, ...);
int   msgBox(HWND hParent, UINT uType, const wchar_t *caption, const wchar_t *msg);
int   msgBoxFmt(HWND hParent, UINT uType, const wchar_t *caption, const wchar_t *msg, ...);
int   runDialog(HINSTANCE hInst, int cmdShow, int dialogId, int iconId, int accelTableId, DLGPROC dialogProc, LPARAM lp);
void  centerOnParent(HWND hDlg);
void  popMenu(HWND hDlg, int popupMenuId, int x, int y, HWND hWndCoordsRelativeTo);
HICON explorerIcon(const wchar_t *fileExtension);
BOOL  openFile(HWND hWnd, const wchar_t *filter, wchar_t *buf, int szBuf);
int   openFiles(HWND hWnd, const wchar_t *filter, wchar_t ***pBuf);


/* Global system font, handled by runDialog(). */
extern HFONT g_hSysFont;
