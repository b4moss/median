//go:build cgo

package pipeline

import (
	"image"

	"github.com/gen2brain/go-fitz"
)

// FitzPDFRenderer renders PDF pages via MuPDF (go-fitz). Requires CGO.
type FitzPDFRenderer struct{}

func NewFitzPDFRenderer() *FitzPDFRenderer { return &FitzPDFRenderer{} }

func (r *FitzPDFRenderer) RenderPage(pdf []byte, page int) (image.Image, error) {
	doc, err := fitz.NewFromMemory(pdf)
	if err != nil {
		return nil, err
	}
	defer doc.Close()
	if page < 0 || page >= doc.NumPage() {
		return nil, ErrInvalidImage
	}
	return doc.Image(page)
}
