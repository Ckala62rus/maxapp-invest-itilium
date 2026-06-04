package api

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"testing"

	"github.com/Ckala62rus/maxapp-invest-itilium/internal/models"
	"github.com/stretchr/testify/require"
)

func TestShrinkImageAttachmentForItilium_largeJPEG(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 2400, 2400))
	for y := 0; y < 2400; y += 4 {
		for x := 0; x < 2400; x += 4 {
			img.Set(x, y, color.RGBA{R: uint8(x % 256), G: uint8(y % 256), B: 128, A: 255})
		}
	}
	var buf bytes.Buffer
	require.NoError(t, jpeg.Encode(&buf, img, &jpeg.Options{Quality: 95}))
	require.Greater(t, buf.Len(), itiliumAttachmentCompressMinBytes)

	out := shrinkImageAttachmentForItilium(models.FileAttachment{
		Filename:    "photo.jpg",
		ContentType: "image/jpeg",
		Data:        buf.Bytes(),
	})
	require.Less(t, len(out.Data), len(buf.Bytes()))
	require.LessOrEqual(t, len(out.Data), itiliumAttachmentMaxBytes)
	require.Equal(t, "image/jpeg", out.ContentType)
}

func TestShrinkImageAttachmentForItilium_smallUnchanged(t *testing.T) {
	data := []byte{1, 2, 3}
	fa := models.FileAttachment{Filename: "x.txt", ContentType: "text/plain", Data: data}
	require.Equal(t, fa, shrinkImageAttachmentForItilium(fa))
}
