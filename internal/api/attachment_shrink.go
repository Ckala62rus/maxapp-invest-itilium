package api

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	_ "image/gif"
	_ "image/png"
	"path/filepath"
	"strings"

	"github.com/Ckala62rus/maxapp-invest-itilium/internal/models"
)

// Лимит тела на nginx перед 1С на itilium_test — по факту ~1 MiB (413 на ~8 MiB JPG).
const itiliumAttachmentMaxBytes = 512 * 1024

const itiliumAttachmentMaxEdge = 1280

var imageFileExt = map[string]bool{
	".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true,
}

var jpegQualities = []int{85, 70, 55, 40, 30}

// prepareItiliumFileAttachments уменьшает крупные фото перед multipart в ITILIUM.
func prepareItiliumFileAttachments(files []models.FileAttachment) []models.FileAttachment {
	if len(files) == 0 {
		return files
	}
	out := make([]models.FileAttachment, len(files))
	for i, fa := range files {
		out[i] = shrinkImageAttachmentForItilium(fa)
	}
	return out
}

func isImageAttachment(fa models.FileAttachment) bool {
	ctype := strings.ToLower(strings.TrimSpace(fa.ContentType))
	if strings.HasPrefix(ctype, "image/") {
		return true
	}
	return imageFileExt[strings.ToLower(filepath.Ext(fa.Filename))]
}

// shrinkImageAttachmentForItilium перекодирует большие изображения в JPEG под лимит nginx 1С.
func shrinkImageAttachmentForItilium(fa models.FileAttachment) models.FileAttachment {
	if !isImageAttachment(fa) {
		return fa
	}
	// С камеры часто 3–10 MiB; пережимаем всё, что больше порога, даже если уже < 1 MiB.
	if len(fa.Data) <= 200*1024 {
		return fa
	}

	img, _, err := image.Decode(bytes.NewReader(fa.Data))
	if err != nil {
		return fa
	}

	scaled := resizeImageMaxEdge(img, itiliumAttachmentMaxEdge)
	data, err := encodeJPEGUnderLimit(scaled, itiliumAttachmentMaxBytes)
	if err != nil || len(data) == 0 {
		return fa
	}
	if len(data) >= len(fa.Data) {
		return fa
	}

	name := fa.Filename
	if ext := filepath.Ext(name); ext != "" {
		name = strings.TrimSuffix(name, ext) + ".jpg"
	} else {
		name = name + ".jpg"
	}

	return models.FileAttachment{
		Filename:    name,
		ContentType: "image/jpeg",
		Data:        data,
	}
}

func resizeImageMaxEdge(src image.Image, maxEdge int) image.Image {
	bounds := src.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	if w <= 0 || h <= 0 {
		return src
	}
	maxSide := w
	if h > maxSide {
		maxSide = h
	}
	if maxSide <= maxEdge {
		return src
	}

	newW := w * maxEdge / maxSide
	newH := h * maxEdge / maxSide
	if newW < 1 {
		newW = 1
	}
	if newH < 1 {
		newH = 1
	}

	dst := image.NewRGBA(image.Rect(0, 0, newW, newH))
	for y := 0; y < newH; y++ {
		sy := bounds.Min.Y + y*h/newH
		for x := 0; x < newW; x++ {
			sx := bounds.Min.X + x*w/newW
			dst.Set(x, y, src.At(sx, sy))
		}
	}
	return dst
}

func encodeJPEGUnderLimit(img image.Image, maxBytes int) ([]byte, error) {
	current := img
	for pass := 0; pass < 4; pass++ {
		for _, q := range jpegQualities {
			var buf bytes.Buffer
			if err := jpeg.Encode(&buf, current, &jpeg.Options{Quality: q}); err != nil {
				return nil, fmt.Errorf("jpeg encode: %w", err)
			}
			if buf.Len() <= maxBytes {
				return buf.Bytes(), nil
			}
		}
		// Ещё уменьшаем сторону, если качество не помогло.
		b := current.Bounds()
		w, h := b.Dx(), b.Dy()
		if w < 320 && h < 320 {
			break
		}
		current = resizeImageMaxEdge(current, max(w, h)/2)
	}

	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, current, &jpeg.Options{Quality: 30}); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
