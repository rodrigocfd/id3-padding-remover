/**
* Handles a memory-mapped file.
*/

#include <Windows.h>

typedef struct {
	HANDLE hFile, hMap;
	void  *pMem;
	int    size;
} FileMap;


FileMap FileMap_new     ();
void    FileMap_free    (FileMap *pfm);
BOOL    FileMap_open    (FileMap *pfm, const wchar_t *path, BOOL readOnly, wchar_t **pErrMsgBuf);
void    FileMap_close   (FileMap *pfm);
void    FileMap_getPtrs (FileMap *pfm, BYTE **pMem, BYTE **pPastEnd);
BOOL    FileMap_truncate(FileMap *pfm, int offset, wchar_t **pErrMsgBuf);


void Uint32Serialize(BYTE *pDest, UINT n, BOOL isBigEndian);
UINT Uint32Unserialize(const BYTE *pSrc, BOOL isBigEndian);
