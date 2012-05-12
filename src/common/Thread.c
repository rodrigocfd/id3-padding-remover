
#include <crtdbg.h>
#include <process.h>
#include "Thread.h"


typedef struct Thread_Sync_Info_ {
	void (*func)(void*); // user callback function
	void *arg;          // user passed argument
} Thread_Sync_Info;

static unsigned int __stdcall _Thread_callback(void *pArgs)
{
	{
		Thread_Sync_Info *tsi = (Thread_Sync_Info*)pArgs;
		tsi->func(tsi->arg); // run user callback function
		free(tsi); // release; allocated at thread setup
	}
	_endthreadex(0); // http://www.codeproject.com/Articles/7732/A-class-to-synchronise-thread-completions/
	return 0;
}

BOOL Thread_RunAsync(void (*runFunc)(void*), void *pArg)
{
	return Thread_RunAsyncTimed(runFunc, pArg, INFINITE); // just run without timeout
}

BOOL Thread_RunAsyncTimed(void (*runFunc)(void*), void *pArg, int msTimeout)
{
	Thread_Sync_Info *tsi;
	HANDLE handle;
	BOOL   ret;

	tsi = malloc(sizeof(Thread_Sync_Info)); // will be freed by _Thread_callback()
	tsi->func = runFunc;
	tsi->arg = pArg;

	handle = (HANDLE)_beginthreadex(NULL, 0, _Thread_callback, tsi, 0, NULL);
	ret = (msTimeout == INFINITE) ? TRUE : (WaitForSingleObject(handle, msTimeout) != WAIT_TIMEOUT);
	CloseHandle(handle);
	return ret;
}

Threads Threads_new(int numThreads)
{
	Threads obj = { 0 };
	obj.handles = malloc(sizeof(HANDLE) * numThreads);
	obj.n = numThreads;
	obj.lastInserted = -1;
	return obj;
}

void Threads_free(Threads *pThreads)
{
	free(pThreads->handles);
	SecureZeroMemory(pThreads, sizeof(Threads));
}

void Threads_add(Threads *pt, void (*runFunc)(void*), void *pArg)
{
	if(pt->lastInserted >= pt->n - 1) // protection against buffer overflow
		MessageBox(0, L"OH NO: Threads_add() called with index beyond limit!", L"Fail", MB_ICONERROR); // shout unpolitely!
	else {
		Thread_Sync_Info *tsi = malloc(sizeof(Thread_Sync_Info)); // will be freed by _Thread_callback()
		tsi->func = runFunc;
		tsi->arg = pArg;
		pt->handles[++pt->lastInserted] = (HANDLE)_beginthreadex(NULL, 0, _Thread_callback, tsi, CREATE_SUSPENDED, NULL);
	}
}

BOOL Threads_runSync(Threads *pt)
{
	return Threads_runSyncTimed(pt, INFINITE); // just run without timeout
}

BOOL Threads_runSyncTimed(Threads *pt, int msTimeout)
{
	BOOL ret;
	int  i;

	_ASSERT(pt->handles);

	for(i = 0; i < pt->n; ++i) {
		if(!pt->handles[i]) { // protection against missing thread objects
			MessageBox(0, L"OH NO: Threads_runSync() called with missing thread object!", L"Fail", MB_ICONERROR); // shout unpolitely!
			ret = FALSE;
			goto bye;
		}
		ResumeThread(pt->handles[i]); // resume each thread, which was created as suspended
	}
	ret = WaitForMultipleObjects(pt->n, pt->handles, TRUE, msTimeout) != WAIT_TIMEOUT; // blocks until all threads return
bye:
	for(i = 0; i < pt->n; ++i)
		if(pt->handles[i])
			CloseHandle(pt->handles[i]);
	return ret;
}
