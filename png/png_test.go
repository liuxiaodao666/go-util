package png

import "testing"

func TestPng(t *testing.T) {
	var callBackMedia CallBackMedia
	callBackMedia.HealthScore = "88"
	callBackMedia.Assessment = "good"

	CreateSVGFile("out.png", callBackMedia)
}
