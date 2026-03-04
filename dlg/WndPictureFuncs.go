//go:build windows

package dlg

import (
	"fmt"
	"id3fit/id3v2"

	"github.com/rodrigocfd/windigo/co"
	"github.com/rodrigocfd/windigo/ui"
	"github.com/rodrigocfd/windigo/win"
)

// Loads the cover art, if due, into the IPicture COM object.
func (me *WndPicture) LoadPicture(tags []*id3v2.Tag) (pixels win.SIZE, nBytes int) {
	picBin, err := me.extractPicBinFromTag(tags)
	if err != nil {
		ui.MsgError(me.wnd.Parent(), "Picture parsing", "", err.Error())
	}

	if picBin == nil {
		return win.SIZE{}, 0 // we don't have a picture to display
	}

	if pixels, err = me.loadBitmap(picBin); err != nil {
		ui.MsgError(me.wnd.Parent(), "Picture parsing", "", err.Error())
	}

	return pixels, len(picBin)
}

func (me *WndPicture) extractPicBinFromTag(tags []*id3v2.Tag) ([]byte, error) {
	apic := id3v2.SameFrameAcrossAllTags("APIC", tags)
	if apic == nil {
		return nil, nil // we don't have a picture to display
	}

	body, ok := apic.Body().(*id3v2.BodyPicture)
	if !ok {
		return nil, fmt.Errorf("APIC frame does not contain BodyPicture body type") // should never happen
	}

	return body.Bin, nil
}

func (me *WndPicture) loadBitmap(picBin []byte) (win.SIZE, error) {
	localOleRel := win.NewOleReleaser()
	defer localOleRel.Release()

	var iFactory *win.IWICImagingFactory
	_ = win.CoCreateInstance(
		localOleRel,
		&co.CLSID_WICImagingFactory,
		nil,
		co.CLSCTX_INPROC_SERVER,
		&iFactory,
	)
	iWicStream, _ := iFactory.CreateStream(localOleRel)
	_ = iWicStream.InitializeFromMemory(picBin)

	iBmpDecoder, err := iFactory.CreateDecoderFromStream(
		localOleRel,
		&iWicStream.IStream,
		nil,
		co.WICDEC_METADATACACHE_OnLoad,
	)
	if err != nil {
		return win.SIZE{}, fmt.Errorf("failed to create BMP decoder: %w", err)
	}

	iFrameDecode, err := iBmpDecoder.GetFrame(localOleRel, 0)
	if err != nil {
		return win.SIZE{}, fmt.Errorf("failed to get frame 0: %w", err)
	}

	iFmtConverter, err := iFactory.CreateFormatConverter(localOleRel)
	if err != nil {
		return win.SIZE{}, fmt.Errorf("failed to create format converter: %w", err)
	}

	err = iFmtConverter.Initialize(
		&iFrameDecode.IWICBitmapSource,
		&co.WIC_PIXELFORMAT_32bppBGRA,
		co.WICBMP_DITHER_None,
		nil,
		0,
		co.WICBMP_PAL_Custom,
	)
	if err != nil {
		return win.SIZE{}, fmt.Errorf("failed to init format converter: %w", err)
	}

	szPixels, err := iFmtConverter.GetSize()
	if err != nil {
		return win.SIZE{}, fmt.Errorf("failed to get sz pixels: %w", err)
	}

	bmi := win.BITMAPINFO{
		BmiHeader: win.BITMAPINFOHEADER{
			Width:       szPixels.Cx,
			Height:      -szPixels.Cy, // top-down
			Planes:      1,
			BitCount:    32,
			Compression: co.BI_RGB,
		},
	}
	bmi.BmiHeader.SetBiSize()

	hBmp, pImgBits, err := win.HDC(0).
		CreateDIBSection(&bmi, co.DIB_COLORS_RGB, win.HFILEMAP(0), 0)
	if err != nil {
		return win.SIZE{}, fmt.Errorf("failed to create DIB section: %w", err)
	}

	me.HBmp.DeleteObject()
	me.HBmp = hBmp         // cache the bitmap
	me.szPixels = szPixels // cache the size

	stride := int(szPixels.Cx) * 4
	bufSize := stride * int(szPixels.Cy)

	err = iFmtConverter.CopyPixels(nil, stride, bufSize, pImgBits)
	if err != nil {
		return win.SIZE{}, fmt.Errorf("failed to copy pixels %w", err)
	}

	return szPixels, nil
}
