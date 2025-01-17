package png

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"testing"
)

func TestDrawText(t *testing.T) {
	// 创建一个新的 RGBA 图像
	img := image.NewRGBA(image.Rect(0, 0, 200, 50))
	// 创建一个简单的 5x5 字体
	font := NewFont("test", 12, 16)

	// 添加一个简单的 "H" 字符的位图数据
	hData := []byte{
		1, 0, 0, 0, 1,
		1, 0, 0, 0, 1,
		1, 1, 1, 1, 1,
		1, 0, 0, 0, 1,
		1, 0, 0, 0, 1,
	}
	font.AddChar('H', hData)

	// 添加一个简单的 "i" 字符的位图数据
	iData := []byte{
		0, 0, 1, 0, 0,
		0, 0, 0, 0, 0,
		0, 0, 1, 0, 0,
		0, 0, 1, 0, 0,
		0, 0, 1, 0, 0,
	}
	font.AddChar('i', iData)

	// 在图像上绘制文本
	err := font.DrawText(img, 10, 10, "Hi", color.RGBA{255, 0, 0, 255}) // 红色文本
	if err != nil {
		t.Fatal(err)
	}

	// 保存为 PNG 文件
	f, err := os.Create("test_output.png")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	if err := png.Encode(f, img); err != nil {
		t.Fatal(err)
	}
}

func TestDrawNumbers(t *testing.T) {
	// 创建一个新的 RGBA 图像
	img := image.NewRGBA(image.Rect(0, 0, 400, 100))
	font := NewFont("numbers", 12, 16)

	// 测试绘制数字
	err := font.DrawText(img, 10, 10, "88", color.RGBA{0, 0, 0, 255}) // 蓝色文本
	if err != nil {
		t.Fatal(err)
	}

	// 保存为 PNG 文件
	f, err := os.Create("numbers.png")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	if err := png.Encode(f, img); err != nil {
		t.Fatal(err)
	}
}
