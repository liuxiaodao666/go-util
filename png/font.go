package png

import (
	"bufio"
	"fmt"
	"image"
	"image/color"
	"os"
	"strconv"
	"strings"
)

// Font 表示一个基本的字体结构
type Font struct {
	Name        string
	Width       int
	Height      int
	Scale       float64 // 添加缩放因子
	Chars       map[rune][]byte
	CharSpacing int
}

// NewFont 创建一个新的字体实例
func NewFont(name string, width, height int) *Font {
	font := &Font{
		Name:        name,
		Width:       width,
		Height:      height,
		Scale:       1.0, // 默认缩放比例为1
		Chars:       make(map[rune][]byte),
		CharSpacing: 1,
	}
	bdfFont, err := parseBDF("spleen-32x64.bdf")
	if err != nil {
		fmt.Printf("Error parsing BDF file: %v\n", err)
		os.Exit(1)
	}

	// generateGoCode(font)

	// 添加所有数字到字体
	for char, data := range bdfFont.Characters {

		if err := font.AddChar(char, data); err != nil {
			fmt.Println(err.Error())
		}
	}

	return font
}

// AddChar 添加字符的位图数据到字体中
func (f *Font) AddChar(char rune, data []byte) error {
	expectedSize := f.Width * f.Height
	if len(data) != expectedSize {
		return fmt.Errorf("invalid character data size: got %d, want %d", len(data), expectedSize)
	}
	f.Chars[char] = data
	return nil
}

// GetChar 获取字符的位图数据
func (f *Font) GetChar(char rune) ([]byte, bool) {
	data, exists := f.Chars[char]
	return data, exists
}

// SetCharSpacing 设置字符间距
func (f *Font) SetCharSpacing(spacing int) {
	f.CharSpacing = spacing
}

// SetScale 设置字体缩放比例
func (f *Font) SetScale(scale float64) {
	if scale > 0 {
		f.Scale = scale
	}
}

// GetScaledSize 获取缩放后的尺寸
func (f *Font) GetScaledSize() (width, height int) {
	return int(float64(f.Width) * f.Scale), int(float64(f.Height) * f.Scale)
}

// DrawText 在指定位置绘制文本
func (f *Font) DrawText(img *image.RGBA, x, y int, text string, color color.Color) error {
	currentX := x
	scaledWidth, _ := f.GetScaledSize() // 只使用scaledWidth
	scaledSpacing := int(float64(f.CharSpacing) * f.Scale)

	for _, char := range text {
		data, exists := f.GetChar(char)
		if !exists {
			continue
		}

		err := f.drawScaledCharacter(img, currentX, y, data, color)
		if err != nil {
			return err
		}

		currentX += scaledWidth + scaledSpacing
	}
	return nil
}

// drawScaledCharacter 在指定位置绘制缩放后的字符
func (f *Font) drawScaledCharacter(img *image.RGBA, x, y int, data []byte, textColor color.Color) error {
	scaledWidth, scaledHeight := f.GetScaledSize()

	if x < 0 || y < 0 || x+scaledWidth > img.Bounds().Max.X || y+scaledHeight > img.Bounds().Max.Y {
		return fmt.Errorf("character position out of bounds")
	}

	// 将输入颜色转换为RGBA
	r, g, b, a := textColor.RGBA()
	tr := uint8(r >> 8)
	tg := uint8(g >> 8)
	tb := uint8(b >> 8)
	ta := uint8(a >> 8)

	// 对每个像素进行缩放和抗锯齿处理
	for dy := 0; dy < scaledHeight; dy++ {
		for dx := 0; dx < scaledWidth; dx++ {
			// 计算原始坐标
			origX := float64(dx) / f.Scale
			origY := float64(dy) / f.Scale

			// 双线性插值
			x0, y0 := int(origX), int(origY)
			x1, y1 := x0+1, y0+1
			if x1 >= f.Width {
				x1 = f.Width - 1
			}
			if y1 >= f.Height {
				y1 = f.Height - 1
			}

			fx := origX - float64(x0)
			fy := origY - float64(y0)

			// 获取周围四个点的值
			v00 := float64(data[y0*f.Width+x0])
			v01 := float64(data[y0*f.Width+x1])
			v10 := float64(data[y1*f.Width+x0])
			v11 := float64(data[y1*f.Width+x1])

			// 双线性插值计算
			intensity := v00*(1-fx)*(1-fy) +
				v01*fx*(1-fy) +
				v10*(1-fx)*fy +
				v11*fx*fy

			// 获取背景色
			bgColor := img.RGBAAt(x+dx, y+dy)

			// 根据强度值计算混合后的颜色
			newR := uint8(float64(bgColor.R)*(1-intensity) + float64(tr)*intensity)
			newG := uint8(float64(bgColor.G)*(1-intensity) + float64(tg)*intensity)
			newB := uint8(float64(bgColor.B)*(1-intensity) + float64(tb)*intensity)
			newA := uint8(float64(bgColor.A)*(1-intensity) + float64(ta)*intensity)

			// 设置新的像素颜色
			img.Set(x+dx, y+dy, color.RGBA{newR, newG, newB, newA})
		}
	}
	return nil
}

type BdfFont struct {
	Name       string
	Size       int
	Width      int
	Height     int
	Characters map[rune][]byte
}

func parseBDF(filename string) (*BdfFont, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	font := &BdfFont{
		Characters: make(map[rune][]byte),
	}

	scanner := bufio.NewScanner(file)
	var currentChar rune
	var bitmap []string
	var collecting bool

	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Fields(line)

		if len(parts) == 0 {
			continue
		}

		switch parts[0] {
		case "FONTBOUNDINGBOX":
			if len(parts) >= 3 {
				font.Width, _ = strconv.Atoi(parts[1])
				font.Height, _ = strconv.Atoi(parts[2])
			}
		case "ENCODING":
			if len(parts) >= 2 {
				encoding, _ := strconv.Atoi(parts[1])
				currentChar = rune(encoding)
			}
		case "BITMAP":
			collecting = true
			bitmap = []string{}
		case "ENDCHAR":
			if collecting {
				font.Characters[currentChar] = convertBitmap(bitmap, font.Width)
				collecting = false
			}
		default:
			if collecting {
				bitmap = append(bitmap, line)
			}
		}
	}

	return font, nil
}

func convertBitmap(bitmap []string, width int) []byte {
	result := make([]byte, 0, len(bitmap)*width)

	for _, line := range bitmap {
		// 将十六进制字符串转换为二进制位
		value, _ := strconv.ParseUint(line, 16, 64)

		// 提取每一位并转换为byte
		for i := width - 1; i >= 0; i-- {
			if value&(1<<uint(i)) != 0 {
				result = append(result, 1)
			} else {
				result = append(result, 0)
			}
		}
	}

	return result
}
