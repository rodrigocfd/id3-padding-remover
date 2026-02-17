//go:build windows

package main

import (
	"id3fit/dlg"
	"runtime"

	"github.com/rodrigocfd/windigo/co"
	"github.com/rodrigocfd/windigo/win"
)

func main() {
	runtime.LockOSThread()

	// go dbgMem()

	win.OleInitialize()
	defer win.OleUninitialize()

	dlg.RunMain()
	// ayy()
}

func ayy() {
	rel := win.NewOleReleaser()
	defer rel.Release()

	var factory *win.IWICImagingFactory
	_ = win.CoCreateInstance(
		rel,
		co.CLSID_WICImagingFactory,
		nil,
		co.CLSCTX_INPROC_SERVER,
		&factory,
	)

	bmpDecoder, _ := factory.CreateDecoderFromFilename(
		rel,
		// "C:\\Temp\\foo.png",
		"c:\\users\\rodrigo\\desktop\\l.png",
		co.GUID_NULL,
		co.GENERIC_READ,
		co.WICDEC_METADATACACHE_OnDemand,
	)

	bmpFrameDec, _ := bmpDecoder.GetFrame(rel, 0)
	sz, _ := bmpFrameDec.GetSize()

	var bi win.BITMAPINFO
	bi.BmiHeader.SetSize()
	bi.BmiHeader.Width = sz.Cx
	bi.BmiHeader.Height = -sz.Cy
	bi.BmiHeader.Planes = 1
	bi.BmiHeader.BitCount = 32
	bi.BmiHeader.Compression = co.BI_RGB

	hdcScreen, _ := win.HWND(0).GetDC()
	defer win.HWND(0).ReleaseDC(hdcScreen)

	hBmp, pImageBits, err := hdcScreen.
		CreateDIBSection(&bi, co.DIB_COLORS_RGB, win.HFILEMAP(0), 0)
	if err != nil {
		println(err.Error())
		return
	}
	defer hBmp.DeleteObject()

	stride := int(sz.Cx) * 4
	szImage := stride * int(sz.Cy)
	if err = bmpFrameDec.CopyPixels(nil, stride, szImage, pImageBits); err != nil {
		println(err.Error())
		return
	}

	//https://faithlife.codes/blog/2008/09/displaying_a_splash_screen_with_c_part_i/
}

// func dbgMem() {
// 	lastFreed := uint64(0)
// 	var stats runtime.MemStats
// 	for {
// 		runtime.ReadMemStats(&stats)
// 		fmt.Printf("GC cycles: %d; Alloc: %s, Next GC: %s; Frees: %d (+%d)\n",
// 			stats.NumGC, win.Str.FmtBytes(stats.HeapAlloc),
// 			win.Str.FmtBytes(stats.NextGC), stats.Frees, stats.Frees-lastFreed)
// 		lastFreed = stats.Frees
// 		time.Sleep(time.Second * 2)
// 	}
// }
