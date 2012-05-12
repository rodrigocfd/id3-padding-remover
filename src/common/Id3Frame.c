
#include "FileMap.h"
#include "Id3Frame.h"


Id3Frame Id3Frame_new()
{
	Id3Frame obj = { 0 };
	return obj;
}

void Id3Frame_free(Id3Frame *pf)
{
	if(pf->type == ID3FRAME_BINARY) {
		if(pf->data.binary.pData)
			free(pf->data.binary.pData); // free data block
	}
	else if(pf->type == ID3FRAME_TEXT) {
		if(pf->data.pText)
			free(pf->data.pText); // free text buffer
	}
	SecureZeroMemory(pf, sizeof(Id3Frame));
}

static void _Id3Frame_parseText(wchar_t **pBuf, const BYTE *pSrcMem, int szData)
{
	if(*pSrcMem == 0x00) // encoding is ISO-8859-1
	{
		int i;

		if(!pSrcMem[szData - 1])
			--szData; // we have a trailing zero, which is useless

		--szData; // minus encoding byte
		*pBuf = malloc(sizeof(wchar_t) * (szData + 1)); // plus terminating null room
		++pSrcMem; // skip encoding, now pointing to 1st string byte

		for(i = 0; i < szData; ++i)
			(*pBuf)[i] = (wchar_t)pSrcMem[i]; // brute force char to wchar_t

		(*pBuf)[szData] = L'\0'; // terminating null
	}
	else if(*pSrcMem == 0x01) // encoding is Unicode UTF-16 with 2-byte BOM
	{
		int  i, numBomBytes, txtLen;
		BOOL isBigEndian;

		++pSrcMem; // skip encoding, now poiting to initial BOM
		if( !(pSrcMem[0] == 0xFE && pSrcMem[1] == 0xFF) && !(pSrcMem[0] == 0xFF && pSrcMem[1] == 0xFE) ) { // validate BOM
			*pBuf = _wcsdup(L"(bad BOM)");
			return;
		}
		--pSrcMem; // get back to encoding byte

		if(!pSrcMem[szData - 2] && !pSrcMem[szData - 1])
			szData -= 2; // we have a trailing zero, which is useless

		numBomBytes = 0; // hoy many BOM individual bytes are present on the whole string (it may be multi-string)
		for(i = 0; i < szData; ++i)
			if(pSrcMem[i] == 0xFE || pSrcMem[i] == 0xFF)
				++numBomBytes;
		szData -= numBomBytes; // BOM bytes won't be stored

		--szData; // minus encoding byte
		txtLen = szData / 2; // half because src chars are 2-byte (wide)
		*pBuf = malloc(sizeof(wchar_t) * (txtLen + 1)); // plus terminating null room
		++pSrcMem; // skip encoding, now poiting to initial BOM

		isBigEndian = (*pSrcMem == 0xFE); // if multi-string, assumes same BOM for all
		pSrcMem += 2; // skip 2-byte BOM
		
		for(i = 0; i < txtLen; ++i, pSrcMem += 2) {
			while(*pSrcMem == 0xFE || *pSrcMem == 0xFF)
				++pSrcMem; // on multi-strings, skip BOM; the individual strings are null separated

			(*pBuf)[i] = isBigEndian ?
				(wchar_t)MAKEWORD(*(pSrcMem + 1), *pSrcMem) : // big endian
				(wchar_t)MAKEWORD(*pSrcMem, *(pSrcMem + 1)); // little endian
		
			(*pBuf)[txtLen] = L'\0'; // terminating null
		}
	}
	else // any other encoding
		*pBuf = _wcsdup(L"(bad encoding)");
}

void Id3Frame_parse(Id3Frame *pf, const BYTE *pSrcMem)
{
	// Frame header structure:
	// [4 byte name] + [4 byte size] + [2 byte flag]

	int i, szData;

	Id3Frame_free(pf);
	*pf = Id3Frame_new(); // fresh start

	for(i = 0; i < 4; ++i)
		pf->name[i] = (wchar_t)*(pSrcMem + i); // brute force char to wchar_t

	szData = Id3Frame_ReadRawDataSize(pSrcMem); // assume non-syncsafed frame size (like Mp3tag writes)
	pSrcMem += 10; // skip 10-byte header, including 2-byte flag

	if(pf->name[0] == L'T') {
		pf->type = ID3FRAME_TEXT;
		_Id3Frame_parseText(&pf->data.pText, pSrcMem, szData); // pass pointer to 1st byte of data block
	}
	else {
		pf->type = ID3FRAME_BINARY;
		pf->data.binary.size = szData; // frame size is the real number of bytes
		pf->data.binary.pData = malloc(sizeof(BYTE) * szData);
		memcpy(pf->data.binary.pData, pSrcMem, sizeof(BYTE) * szData); // simply store raw data
	}	
}

const wchar_t* Id3Frame_getText(Id3Frame *pf)
{
	if(pf->type != ID3FRAME_TEXT) return NULL;
	return pf->data.pText;
}

int Id3Frame_getDataSize(Id3Frame *pf)
{
	if(pf->type != ID3FRAME_BINARY) return 0;
	return pf->data.binary.size;
}

const BYTE* Id3Frame_getData(Id3Frame *pf)
{
	if(pf->type != ID3FRAME_BINARY) return NULL;
	return pf->data.binary.pData;
}

int SynchsafeEnc(int in)
{
	int out, mask = 0x7F;
	while(mask ^ 0x7FFFFFFF) {
		out = in & ~mask;
		out <<= 1;
		out |= in & mask;
		mask = ((mask + 1) << 8) - 1;
		in = out;
	}
	return out; // encoded synchsafe int32
}

int SynchsafeDec(int in)
{
	int out = 0, mask = 0x7F000000;
	while(mask) {
		out >>= 1;
		out |= in & mask;
		mask >>= 8;
	}
	return out; // decoded synchsafe int32
}
