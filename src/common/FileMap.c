
#include <crtdbg.h>
#include "FileMap.h"


FileMap FileMap_new()
{
	FileMap obj = { 0 };
	return obj;
}

void FileMap_free(FileMap *pfm)
{
	FileMap_close(pfm);
}

BOOL FileMap_open(FileMap *pfm, const wchar_t *path, BOOL readOnly, wchar_t **pErrMsgBuf)
{
	FileMap_close(pfm); // make sure everything was properly cleaned up

	// Open the file for reading or writing.
	pfm->hFile = CreateFile(path,
		GENERIC_READ | (readOnly ? 0 : GENERIC_WRITE),
		readOnly ? FILE_SHARE_READ : 0,
		NULL, readOnly ? OPEN_EXISTING : OPEN_ALWAYS, 0, 0);
	
	if(pfm->hFile == INVALID_HANDLE_VALUE) {
		FileMap_close(pfm);
		if(pErrMsgBuf) // user must free buf
			*pErrMsgBuf = _wcsdup(readOnly ? L"Could not open file for reading." : L"Could not open file for writing.");
		return FALSE;
	}

	// Map file into memory.
	pfm->hMap = CreateFileMapping(pfm->hFile, NULL, readOnly ? PAGE_READONLY : PAGE_READWRITE, 0, 0, NULL);
	if(!pfm->hMap) {
		FileMap_close(pfm);
		if(pErrMsgBuf)
			*pErrMsgBuf = _wcsdup(L"Could not create file mapping."); // user must free buf
		return FALSE;
	}

	// Get pointer to data block.
	pfm->pMem = MapViewOfFile(pfm->hMap, readOnly ? FILE_MAP_READ : FILE_MAP_WRITE, 0, 0, 0);
	if(!pfm->pMem) {
		FileMap_close(pfm);
		if(pErrMsgBuf)
			*pErrMsgBuf = _wcsdup(L"Could not map view of file."); // user must free buf
		return FALSE;
	}

	// Keep file size.
	pfm->size = GetFileSize(pfm->hFile, NULL);

	return TRUE;
}

void FileMap_close(FileMap *pfm)
{
	if(pfm->pMem) UnmapViewOfFile(pfm->pMem);
	if(pfm->hMap) CloseHandle(pfm->hMap);
	if(pfm->hFile) CloseHandle(pfm->hFile);
	SecureZeroMemory(pfm, sizeof(FileMap));
}

void FileMap_getPtrs(FileMap *pfm, BYTE **pMem, BYTE **pPastEnd)
{
	_ASSERT(pfm->hFile && pfm->hMap && pfm->pMem);
	if(pMem) *pMem = (BYTE*)pfm->pMem;
	if(pPastEnd) *pPastEnd = (BYTE*)pfm->pMem + pfm->size;
}

BOOL FileMap_truncate(FileMap *pfm, int offset, wchar_t **pErrMsgBuf)
{
	// This function will fail if file was opened as read-only.

	_ASSERT(pfm->hFile && pfm->hMap && pfm->pMem);

	if(!offset) // because it fails at zero
		offset = 1;
	else if(offset > 0 && offset > pfm->size) // cannot truncate beyond file
		offset = pfm->size;
	else if(offset < 0 && abs(offset) > pfm->size) // cannot truncate before first byte
		offset = 1;

	// Unmap file, but keep it open.
	UnmapViewOfFile(pfm->pMem);
	CloseHandle(pfm->hMap);

	// Truncate file; negative offset cuts from end to beginning.
	if(SetFilePointer(pfm->hFile, offset, NULL, offset > 0 ? FILE_BEGIN : FILE_END) == INVALID_SET_FILE_POINTER) {
		FileMap_close(pfm);
		if(pErrMsgBuf)
			*pErrMsgBuf = _wcsdup(L"Could not set file pointer position."); // user must free buf
		return FALSE;
	}

	if(!SetEndOfFile(pfm->hFile)) {
		FileMap_close(pfm);
		if(pErrMsgBuf)
			*pErrMsgBuf = _wcsdup(L"Could not set new end of file."); // user must free buf
		return FALSE;
	}
	SetFilePointer(pfm->hFile, 0, NULL, FILE_BEGIN); // rewind

	// Remapping into memory.
	pfm->hMap = CreateFileMapping(pfm->hFile, NULL, PAGE_READWRITE, 0, 0, NULL);
	if(!pfm->hMap) {
		FileMap_close(pfm);
		if(pErrMsgBuf)
			*pErrMsgBuf = _wcsdup(L"Could not recreate file mapping."); // user must free buf
		return FALSE;
	}

	// Get new pointer to data block, old one just became invalid!
	pfm->pMem = MapViewOfFile(pfm->hMap, FILE_MAP_WRITE, 0, 0, 0);
	if(!pfm->pMem) {
		FileMap_close(pfm);
		if(pErrMsgBuf)
			*pErrMsgBuf = _wcsdup(L"Could not remap view of file."); // user must free buf
		return FALSE;
	}

	// Calculate new file size.
	if(offset > 0) pfm->size = offset;
	else pfm->size += offset;

	return TRUE;
}

void Uint32Serialize(BYTE *pDest, UINT n, BOOL isBigEndian)
{
	if(isBigEndian) {
		pDest[0] = (n & 0xFF000000) >> 24;
		pDest[1] = (n & 0xFF0000) >> 16;
		pDest[2] = (n & 0xFF00) >> 8;
		pDest[3] = n & 0xFF;
	}
	else {
		pDest[0] = n & 0xFF;
		pDest[1] = (n & 0xFF00) >> 8;
		pDest[2] = (n & 0xFF0000) >> 16;
		pDest[3] = (n & 0xFF000000) >> 24;
	}
}

UINT Uint32Unserialize(const BYTE *pSrc, BOOL isBigEndian)
{
	return isBigEndian ?
		(((BYTE*)(pSrc))[0] << 24) | (((BYTE*)(pSrc))[1] << 16) | (((BYTE*)(pSrc))[2] << 8) | ((BYTE*)(pSrc))[3] :
		((BYTE*)(pSrc))[0] | (((BYTE*)(pSrc))[1] << 8) | (((BYTE*)(pSrc))[2] << 16) | (((BYTE*)(pSrc))[3] << 24);
}
