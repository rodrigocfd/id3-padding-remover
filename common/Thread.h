/**
* Some thread wrapper routines, to work with both sync and asynchronous threads,
* taking care of the all the required setup and cleanup.
*/

#include <Windows.h>


/* Asynchronous thread. */
BOOL Thread_RunAsync     (void (*runFunc)(void*), void *pArg);
BOOL Thread_RunAsyncTimed(void (*runFunc)(void*), void *pArg, int msTimeout);


/* Synchronous threads. */
typedef struct {
	HANDLE *handles;
	int     n, lastInserted;
} Threads;

Threads Threads_new         (int numThreads);
void    Threads_free        (Threads *pt);
void    Threads_add         (Threads *pt, void (*runFunc)(void*), void *pArg);
BOOL    Threads_runSync     (Threads *pt);
BOOL    Threads_runSyncTimed(Threads *pt, int msTimeout);
