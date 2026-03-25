package pdf_test

import (
	"testing"
	_ "unsafe"

	"github.com/stretchr/testify/assert"

	"github.com/opendataloader-project/opendataloader-pdf-go/internal/entities"
	internalpdf "github.com/opendataloader-project/opendataloader-pdf-go/internal/pdf"
)

//go:linkname getColor github.com/opendataloader-project/opendataloader-pdf-go/internal/pdf.getColor
func getColor(objType entities.ObjectType) [3]float64

func TestGetColor(t *testing.T) {
	assert.Equal(t, [3]float64{0, 0, 1}, getColor(entities.ObjectTypeHeading))
	assert.Equal(t, [3]float64{0, 1, 0}, getColor(entities.ObjectTypeList))
	assert.Equal(t, [3]float64{1, 0, 1}, getColor(entities.ObjectTypeTable))
}

func TestPDFLayerValues(t *testing.T) {
	assert.Equal(t, internalpdf.PDFLayer("content"), internalpdf.PDFLayerContent)
	assert.Equal(t, internalpdf.PDFLayer("table cells"), internalpdf.PDFLayerTableCells)
}
