package main

const (
	APP_TITLE = "ID3 Fit"
	ICO_MAIN  = 101
)

const (
	LST_FILES = iota + 1001
	LST_FRAMES
)

const (
	MNU_MAIN = iota + 200
	MNU_OPEN
	MNU_DELETE
	MNU_REM_PAD
	MNU_REM_RG
	MNU_REM_RG_PIC
	MNU_PREFIX_YEAR
	MNU_CLEAR_DIACR
	MNU_ABOUT
)

const (
	TIMER_LSTFILES uintptr = 100
)
