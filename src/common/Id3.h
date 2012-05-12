/**
* Handles an ID3v2 tag.
* http://www.id3.org/id3v2.3.0
*/

#include "FileMap.h"
#include "Id3Frame.h"

typedef struct Id3_ {
	FileMap fm;
} Id3;


Id3  Id3_new          ();
void Id3_free         (Id3 *pid3);
BOOL Id3_open         (Id3 *pid3, const wchar_t *mp3file, wchar_t **pErrMsgBuf);
void Id3_close        (Id3 *pid3);
int  Id3_totalTagSize (Id3 *pid3);
int  Id3_mp3TailSize  (Id3 *pid3);
int  Id3_paddingSize  (Id3 *pid3);
BOOL Id3_removePadding(Id3 *pid3);
int  Id3_countFrames  (Id3 *pid3);
int  Id3_getFrames    (Id3 *pid3, Id3Frame **pBuf, wchar_t **pErrMsgBuf);
