package main

// Command IDs used in the application.

const (
	MNU_OPEN int = iota + 1001
	MNU_DELETE
	MNU_REMPAD
	MNU_REMRG
	MNU_REMRGPIC
	MNU_ABOUT

	LST_FILES
	LST_VALUES
)

const (
	TIMER_LSTFILES uintptr = iota + 100
)
