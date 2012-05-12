
#include <Windows.h>

#define ID3FRAME_UNDEFINED 0
#define ID3FRAME_BINARY    1
#define ID3FRAME_TEXT      2

typedef struct Id3Frame_ {
	wchar_t name[5];
	BYTE    type;
	union {
		wchar_t *pText;
		struct { int size; BYTE *pData; } binary;
	} data;
} Id3Frame;


Id3Frame       Id3Frame_new        ();
void           Id3Frame_free       (Id3Frame *pf);
void           Id3Frame_parse      (Id3Frame *pf, const BYTE *pSrcMem);
const wchar_t* Id3Frame_getText    (Id3Frame *pf);
int            Id3Frame_getDataSize(Id3Frame *pf);
const BYTE*    Id3Frame_getData    (Id3Frame *pf);


/* Data size of frame directly read from pointer;
   assumes not-syncsafed, like Mp3tag writes. */
#define Id3Frame_ReadRawDataSize(pMem) Uint32Unserialize(pMem + 4, TRUE)


/* Global helper functions. */
int SynchsafeEnc(int in);
int SynchsafeDec(int in);
