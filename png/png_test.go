package png

import (
	"testing"
)

func TestPng(t *testing.T) {
	var callBackMedia CallBackMedia
	callBackMedia.HealthScore = "70"
	callBackMedia.Assessment = "excellent"

	CreateSVGFile("out.png", callBackMedia)
}
