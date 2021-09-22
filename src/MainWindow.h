
#include <core/MainDialog.h>
#include <core/ImageList.h>
#include <core/Menu.h>

class MainWindow final : public core::MainDialog {
public:
	virtual ~MainWindow();
	MainWindow();

private:
	core::ImageList iconsList;
	core::Menu appMenu;

	virtual INT_PTR dialogProc(UINT msg, WPARAM wp, LPARAM lp) override;
	void onInitDialog();
	void onFilesOpen();
	void onFilesAbout();
};
