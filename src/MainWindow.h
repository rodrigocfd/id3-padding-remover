
#include <core/MainDialog.h>

class MainWindow final : public core::MainDialog {
public:
	MainWindow();

private:
	virtual INT_PTR dialogProc(UINT msg, WPARAM wp, LPARAM lp) override;
	void onInitDialog();
};
