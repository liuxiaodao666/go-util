package png

import "testing"

func TestPng(t *testing.T) {
	var callBackMedia CallBackMedia
	callBackMedia.HealthScore = "96"
	callBackMedia.Assessment = "excellent"

	CreateSVGFile("out.png", callBackMedia)
}
