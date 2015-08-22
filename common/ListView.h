/**
* ListView control shorthand routines and macros.
* The idea is, when dealing with a ListView control, use only these, and don't
* call directly the system routines/macros.
*/

#include <Windows.h>
#include <CommCtrl.h>


#define ListView_fullRowSel(hlist)                  ListView_SetExtendedListViewStyle(hlist, LVS_EX_FULLROWSELECT)
#define ListView_selCount(hlist)                    ListView_GetSelectedCount(hlist)
#define ListView_count(hlist)                       ListView_GetItemCount(hlist)
#define ListView_selAllItems(hlist)                 ListView_SetItemState(hlist, -1, LVIS_SELECTED, LVIS_SELECTED)
#define ListView_getNextSel(hlist, i)               ListView_GetNextItem(hlist, i, LVNI_SELECTED)
#define ListView_setText(hlist, i, col, text)       ListView_SetItemText(hlist, i, col, (wchar_t*)(text))
#define ListView_getText(hlist, i, col, buf, bufsz) ListView_GetItemText(hlist, i, col, buf, bufsz)

void ListView_addColumn  (HWND hList, const wchar_t *caption, int cx);
void ListView_fitColumn  (HWND hList, int iCol);
int  ListView_addItem    (HWND hList, const wchar_t *caption, int iconIdx);
int  ListView_pushIcon   (HWND hList, int iconId);
int  ListView_pushSysIcon(HWND hList, const wchar_t *fileExtension);
BOOL ListView_itemExists (HWND hList, const wchar_t *caption);
void ListView_delSelItems(HWND hList);
void ListView_setTextFmt (HWND hList, int i, int col, const wchar_t *fmt, ...);
int  ListView_popMenu    (HWND hList, int popupMenuId, BOOL popEvenWithoutItem);
