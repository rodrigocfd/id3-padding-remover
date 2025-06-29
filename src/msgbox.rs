//! Standard message boxes to retrieve user input.

use winsafe::{self as w, co, prelude::*};

/// Informational message.
pub fn info(
	parent: &impl GuiParent,
	title: &str,
	main: Option<&str>,
	text: &str,
) -> w::HrResult<()> {
	ok(parent, title, main, text, false)
}

/// Error message.
pub fn err(
	parent: &impl GuiParent,
	title: &str,
	main: Option<&str>,
	text: &str,
) -> w::HrResult<()> {
	ok(parent, title, main, text, true)
}

fn ok(
	parent: &impl GuiParent,
	title: &str,
	main: Option<&str>,
	text: &str,
	is_err: bool,
) -> w::HrResult<()> {
	w::TaskDialogIndirect(&w::TASKDIALOGCONFIG {
		hwnd_parent: Some(parent.hwnd()),
		window_title: Some(title),
		main_instruction: main,
		main_icon: w::IconIdTd::Td(if is_err {
			co::TD_ICON::ERROR
		} else {
			co::TD_ICON::INFORMATION
		}),
		common_buttons: co::TDCBF::OK,
		flags: co::TDF::ALLOW_DIALOG_CANCELLATION | co::TDF::POSITION_RELATIVE_TO_WINDOW,
		content: Some(text),
		..Default::default()
	})?;

	Ok(())
}

/// OK/Cancel question.
#[must_use]
pub fn ask(
	parent: &impl GuiParent,
	title: &str,
	main: Option<&str>,
	text: &str,
	ok_text: &str,
) -> w::HrResult<bool> {
	let (res, _, _) = w::TaskDialogIndirect(&w::TASKDIALOGCONFIG {
		hwnd_parent: Some(parent.hwnd()),
		window_title: Some(title),
		main_instruction: main,
		main_icon: w::IconIdTd::Td(co::TD_ICON::WARNING),
		common_buttons: co::TDCBF::CANCEL,
		buttons: &[(co::DLGID::OK.into(), ok_text)],
		flags: co::TDF::ALLOW_DIALOG_CANCELLATION | co::TDF::POSITION_RELATIVE_TO_WINDOW,
		content: Some(text),
		..Default::default()
	})?;

	Ok(res == co::DLGID::OK)
}
