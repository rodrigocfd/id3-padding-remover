
#include "Id3.h"
#include "util.h"


Id3 Id3_new()
{
	Id3 obj = { 0 };
	obj.fm = FileMap_new(); // create mapped file object
	return obj;
}

void Id3_free(Id3 *pid3)
{
	Id3_close(pid3);
}

BOOL Id3_open(Id3 *pid3, const wchar_t *mp3file, wchar_t **pErrMsgBuf)
{
	if(!FileMap_open(&pid3->fm, mp3file, FALSE, pErrMsgBuf)) // user must free buf
		return FALSE;
	return TRUE;
}

void Id3_close(Id3 *pid3)
{
	FileMap_free(&pid3->fm); // free mapped file object
}

int Id3_totalTagSize(Id3 *pid3)
{
	BYTE *pMem = 0; // pointer to 1st byte of file
	FileMap_getPtrs(&pid3->fm, &pMem, NULL);
	if(memcmp(pMem, (BYTE*)"ID3", 3)) // data block doesn't begin with "ID3"?
		return 0; // no tag, size is zero
	return SynchsafeDec( Uint32Unserialize(pMem + 6, TRUE) ) + 10; // include 10-byte header
}

WORD Id3_tagVersion(Id3 *pid3)
{
	BYTE *pMem = 0; // pointer to 1st byte of file
	FileMap_getPtrs(&pid3->fm, &pMem, NULL);
	return MAKEWORD(pMem[4], pMem[3]); // (lo,hi) like (0,3) for v2.3.0
}

int Id3_mp3TailSize(Id3 *pid3)
{
	BYTE *pPast = 0, *pRun;

	FileMap_getPtrs(&pid3->fm, NULL, &pPast);
	pRun = pPast - 1;
	while(*pRun == '\0') --pRun;
	return pPast - pRun - 1; // useless zero bytes eventually found at the end of MP3 file
}

int Id3_paddingSize(Id3 *pid3)
{
	BYTE *pMem = 0; // pointer to 1st byte of file
	int totalTagSize, off;

	FileMap_getPtrs(&pid3->fm, &pMem, NULL);
	totalTagSize = Id3_totalTagSize(pid3); // including 10-byte tag header; including padding
	off = 10; // skip 10-byte tag header

	while(off < totalTagSize) {
		if(!within(pMem[off], '0', '9') && !within(pMem[off], 'A', 'Z'))
			break; // probably entered a padding region

		// Advance to 1st byte of next frame; include 10-byte frame header.
		// Assume non-syncsafed frame size (like Mp3tag writes).
		off += Id3Frame_ReadRawDataSize(pMem + off) + 10;
	}

	return totalTagSize - off; // length of padding
}

BOOL Id3_removePadding(Id3 *pid3)
{
	BYTE *pMem = 0, *pPast = 0;
	int totalTagSize, padLen, tailSize;

	FileMap_getPtrs(&pid3->fm, &pMem, &pPast);
	totalTagSize = Id3_totalTagSize(pid3); // including 10-byte tag header; including padding
	if(!totalTagSize)
		return FALSE; // no tag found

	padLen = Id3_paddingSize(pid3);
	if(padLen) {
		// Write tag size minus the padding length.
		Uint32Serialize(pMem + 6,
			SynchsafeEnc(totalTagSize - 10 - padLen), TRUE); // do not count the 10-byte tag header

		// Move the whole MP3 memory block back, over the padding room.
		memmove(pMem + totalTagSize - padLen,
			pMem + totalTagSize,
			pPast - pMem - totalTagSize);
	}

	tailSize = Id3_mp3TailSize(pid3); // useless zero bytes eventually found at end of MP3
	if(padLen + tailSize)
		return FileMap_truncate(&pid3->fm, -(padLen + tailSize)); // truncate file
	return TRUE;
}

int Id3_countFrames(Id3 *pid3)
{
	BYTE *pMem = 0;
	int totalTagSize = Id3_totalTagSize(pid3);
	int numFrames = 0, off = 10; // skip 10-byte header

	FileMap_getPtrs(&pid3->fm, &pMem, NULL);
	
	while(off < totalTagSize) {
		if(!within(pMem[off], '0', '9') && !within(pMem[off], 'A', 'Z'))
			break; // probably entered a padding region

		off += Id3Frame_ReadRawDataSize(pMem + off) + 10; // advance to 1st byte of next frame (include 10-byte frame header)
		++numFrames;
	}

	return numFrames;
}

int Id3_getFrames(Id3 *pid3, Id3Frame **pBuf, wchar_t **pErrMsgBuf)
{
	BYTE *pMem = 0, *pPast = 0;
	int totalTagSize, numFrames = 0;

	FileMap_getPtrs(&pid3->fm, &pMem, &pPast);
	if(!( totalTagSize = Id3_totalTagSize(pid3) )) { // including 10-byte header; including padding
		if(pErrMsgBuf)
			*pErrMsgBuf = _wcsdup(L"No tag found.");
		return -1;
	}

	if(Id3_tagVersion(pid3) != MAKEWORD(0, 3)) { // not v2.3.0?
		if(pErrMsgBuf)
			*pErrMsgBuf = allocfmt(L"Unhandled tag version: 2.%d.%d (only 2.3.0 is supported).",
				HIBYTE(Id3_tagVersion(pid3)), LOBYTE(Id3_tagVersion(pid3)) ); // user must free buf
		return -1;
	}

	if(pMem[5] & 0x80) { // unsynchronisation (7th) bit is set?
		if(pErrMsgBuf)
			*pErrMsgBuf = _wcsdup(L"Unsynchronisated tags are not supported."); // user must free buf
		return -1;
	}

	// Alloc return buffer.
	numFrames = Id3_countFrames(pid3);
	*pBuf = malloc(sizeof(Id3Frame) * numFrames);

	// Load all frames into memory.
	{
		int i, off = 10; // skip 10-byte header
		for(i = 0; i < numFrames; ++i) {
			Id3Frame *pCurFrame = &( (*pBuf)[i] );
			*pCurFrame = Id3Frame_new(); // init frame object
			Id3Frame_parse(pCurFrame, pMem + off);
			off += Id3Frame_ReadRawDataSize(pMem + off) + 10; // advance to 1st byte of next frame (include 10-byte frame header)
		}
	}

	return numFrames; // return buffer must be freed
}
