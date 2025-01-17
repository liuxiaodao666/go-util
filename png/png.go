package png

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"strconv"

	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/goregular"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

// getFanShapeColor 根据评估等级返回颜色
func getFanShapeColor(assessment string) [3]float64 {
	colorMap := map[string][3]float64{
		"excellent": {0.70, 0.87, 0.40}, //"#b4de66", RGB(0.70,0.87,0.40)
		"good":      {0.96, 0.78, 0.05}, //"#f5c70f", RGB(0.96,0.78,0.05)
		"normal":    {0.96, 0.61, 0.24}, //"#f59c3d", RGB(0.96,0.61,0.24)
		"poor":      {0.96, 0.42, 0.41}, //"#f76d6a", RGB(0.96,0.42,0.41)
	}
	return colorMap[assessment]
}

type CallBackMedia struct {
	HealthScore string `json:"health_score"`
	Assessment  string `json:"assessment"`
}

func saveDebugImage(img *image.RGBA, name string) {
	f, _ := os.Create(name)
	defer f.Close()
	png.Encode(f, img)
}

// createSVGFile 创建 SVG 文件并返回文件路径
func CreateSVGFile(dirPath string, callBackMedia CallBackMedia) (string, error) {
	var fanShapeColor [3]float64
	var scoreFloat float64
	var fanShapeEndAngle float64

	// 将字符串转换为浮点数
	scoreFloat, err := strconv.ParseFloat(callBackMedia.HealthScore, 64)
	if err != nil {
		return "", fmt.Errorf("Error converting score to float: %v", err)
	}
	//fmt.Printf("Parsed score as float: %.1f\n", scoreFloat)

	//计算扇形结束角度
	fanShapeEndAngle = 360*(scoreFloat/100) + 270
	//扇形颜色
	fanShapeColor = getFanShapeColor(callBackMedia.Assessment)
	//fmt.Printf("Fan shape color: %s\n", fanShapeColor)

	drawPng(fanShapeColor, fanShapeEndAngle, scoreFloat, dirPath)

	return dirPath, nil
}

// 添加 alpha 混合函数
func alphaBlend(dst, src color.Color) color.RGBA {
	dr, dg, db, da := dst.RGBA()
	sr, sg, sb, sa := src.RGBA()

	// 转换到 0-255 范围
	dr >>= 8
	dg >>= 8
	db >>= 8
	da >>= 8
	sr >>= 8
	sg >>= 8
	sb >>= 8
	sa >>= 8

	// alpha 混合计算
	a := sa + da*(255-sa)/255
	if a == 0 {
		return color.RGBA{0, 0, 0, 0}
	}

	r := (sr*sa + dr*da*(255-sa)/255) / a
	g := (sg*sa + dg*da*(255-sa)/255) / a
	b := (sb*sa + db*da*(255-sa)/255) / a

	return color.RGBA{uint8(r), uint8(g), uint8(b), uint8(a)}
}

// 修改 drawPng 函数
func drawPng(fanShapeColor [3]float64, fanShapeEndAngle, scoreFloat float64, pngName string) {
	// 创建一个新的 RGBA 图像
	img := image.NewRGBA(image.Rect(0, 0, 300, 200))

	// // 1. 填充白色背景
	// for y := 0; y < img.Bounds().Dy(); y++ {
	// 	for x := 0; x < img.Bounds().Dx(); x++ {
	// 		img.Set(x, y, color.White)
	// 	}
	// }
	// saveDebugImage(img, "0_background.png")

	// 2. 创建临时缓冲区用于每个图层
	tempImg := image.NewRGBA(img.Bounds())
	copy(tempImg.Pix, img.Pix)

	// 3. 绘制灰色背景圆
	drawCircle(tempImg, 150, 100, 90, color.RGBA{237, 237, 237, 255})
	// 混合到主图层
	blendLayers(img, tempImg)
	saveDebugImage(img, "1_gray_circle.png")

	// 4. 绘制扇形
	clear(tempImg)
	drawSector(tempImg, 150, 100, 90, 270, fanShapeEndAngle, fanShapeColor)
	// 混合到主图层
	blendLayers(img, tempImg)
	saveDebugImage(img, "2_sector.png")

	// 5. 绘制中心白色圆
	clear(tempImg)
	drawCircle(tempImg, 150, 100, 50, color.RGBA{255, 255, 255, 245})
	blendLayers(img, tempImg)
	saveDebugImage(img, "3_white_circle.png")

	// 6. 绘制文字
	drawText(img, scoreFloat, 150, 100)
	saveDebugImage(img, "4_text.png")

	// 保存最终结果
	f, _ := os.Create(pngName)
	defer f.Close()
	png.Encode(f, img)
}

// 清空临时缓冲区
func clear(img *image.RGBA) {
	for i := range img.Pix {
		img.Pix[i] = 0
	}
}

// 混合两个图层
func blendLayers(dst, src *image.RGBA) {
	bounds := dst.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			dstColor := dst.RGBAAt(x, y)
			srcColor := src.RGBAAt(x, y)
			if srcColor.A > 0 { // 只处理非透明像素
				blended := alphaBlend(dstColor, srcColor)
				dst.Set(x, y, blended)
			}
		}
	}
}

// 添加抗锯齿相关常量
const (
	kappa      = 0.5522847498
	ssaaScale  = 2      // 超采样倍数
	bezierStep = 0.0005 // 贝塞尔曲线采样精度
)

// 修改 drawCircle 函数，为小圆提供更好的抗锯齿效果
func drawCircle(img *image.RGBA, centerX, centerY, radius int, c color.Color) {
	cr, cg, cb, ca := c.RGBA()
	cr >>= 8
	cg >>= 8
	cb >>= 8
	ca >>= 8

	// 增加采样倍数
	localScale := ssaaScale * 4 // 增加采样倍数

	ssaaImg := image.NewRGBA(image.Rect(0, 0, img.Bounds().Dx()*localScale, img.Bounds().Dy()*localScale))
	sCenterX := centerX * localScale
	sCenterY := centerY * localScale
	sRadius := radius * localScale

	// 增加边缘柔化范围
	edgeWidth := float64(localScale) * 2.0 // 增加边缘过渡区域

	for y := -sRadius - int(edgeWidth); y <= sRadius+int(edgeWidth); y++ {
		for x := -sRadius - int(edgeWidth); x <= sRadius+int(edgeWidth); x++ {
			dist := math.Sqrt(float64(x*x + y*y))

			// 使用更平滑的过渡函数
			alpha := 1.0
			if dist > float64(sRadius) {
				alpha = 0.0
			} else if dist > float64(sRadius)-edgeWidth {
				// 使用更平滑的过渡
				t := (float64(sRadius) - dist) / edgeWidth
				alpha = smoothstep(t)
			}

			if alpha > 0 {
				// 应用颜色
				ssaaImg.Set(sCenterX+x, sCenterY+y, color.RGBA{
					R: uint8(float64(cr) * alpha),
					G: uint8(float64(cg) * alpha),
					B: uint8(float64(cb) * alpha),
					A: uint8(float64(ca) * alpha),
				})
			}
		}
	}

	// 使用高斯模糊进行平滑
	gaussianBlur(ssaaImg, 1.5)

	// 缩放回原始大小
	downscaleWithRadius(ssaaImg, img, localScale)
}

// 添加高斯模糊函数
func gaussianBlur(img *image.RGBA, sigma float64) {
	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()

	// 创建临时缓冲区
	temp := image.NewRGBA(bounds)

	// 计算高斯核大小
	kernelSize := int(sigma * 3)
	kernel := make([]float64, kernelSize*2+1)
	for i := range kernel {
		x := float64(i - kernelSize)
		kernel[i] = math.Exp(-(x * x) / (2 * sigma * sigma))
	}

	// 水平方向模糊
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			var r, g, b, a float64
			var sum float64

			for i := -kernelSize; i <= kernelSize; i++ {
				px := x + i
				if px < 0 {
					px = 0
				}
				if px >= w {
					px = w - 1
				}

				weight := kernel[i+kernelSize]
				pixel := img.RGBAAt(px, y)
				r += float64(pixel.R) * weight
				g += float64(pixel.G) * weight
				b += float64(pixel.B) * weight
				a += float64(pixel.A) * weight
				sum += weight
			}

			temp.Set(x, y, color.RGBA{
				uint8(r / sum),
				uint8(g / sum),
				uint8(b / sum),
				uint8(a / sum),
			})
		}
	}

	// 垂直方向模糊
	for x := 0; x < w; x++ {
		for y := 0; y < h; y++ {
			var r, g, b, a float64
			var sum float64

			for i := -kernelSize; i <= kernelSize; i++ {
				py := y + i
				if py < 0 {
					py = 0
				}
				if py >= h {
					py = h - 1
				}

				weight := kernel[i+kernelSize]
				pixel := temp.RGBAAt(x, py)
				r += float64(pixel.R) * weight
				g += float64(pixel.G) * weight
				b += float64(pixel.B) * weight
				a += float64(pixel.A) * weight
				sum += weight
			}

			img.Set(x, y, color.RGBA{
				uint8(r / sum),
				uint8(g / sum),
				uint8(b / sum),
				uint8(a / sum),
			})
		}
	}
}

// 改进的平滑过渡函数
func smoothstep(x float64) float64 {
	if x <= 0 {
		return 0
	}
	if x >= 1 {
		return 1
	}
	return x * x * x * (x*(x*6-15) + 10)
}

// 根据不同的采样倍数进行缩放
func downscaleWithRadius(ssaa *image.RGBA, dst *image.RGBA, scale int) {
	bounds := dst.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			var r, g, b, a float64

			// 采样超采样缓冲区的对应像素
			for dy := 0; dy < scale; dy++ {
				for dx := 0; dx < scale; dx++ {
					sx := x*scale + dx
					sy := y*scale + dy
					c := ssaa.RGBAAt(sx, sy)
					r += float64(c.R)
					g += float64(c.G)
					b += float64(c.B)
					a += float64(c.A)
				}
			}

			// 计算平均值
			scaleSquared := float64(scale * scale)
			dst.Set(x, y, color.RGBA{
				uint8(r / scaleSquared),
				uint8(g / scaleSquared),
				uint8(b / scaleSquared),
				uint8(a / scaleSquared),
			})
		}
	}
}

// 绘制带抗锯齿的贝塞尔曲线
func drawAntialiasedBezier(img *image.RGBA, x0, y0, x1, y1, x2, y2, x3, y3 int, c color.Color) {
	r, g, b, a := c.RGBA()
	for t := 0.0; t <= 1.0; t += bezierStep {
		xt := cubic(float64(x0), float64(x1), float64(x2), float64(x3), t)
		yt := cubic(float64(y0), float64(y1), float64(y2), float64(y3), t)

		// 对点进行抗锯齿处理
		drawAntialiasedPoint(img, int(xt), int(yt),
			uint8(r>>8), uint8(g>>8), uint8(b>>8), uint8(a>>8))
	}
}

// 绘制带抗锯齿的点
func drawAntialiasedPoint(img *image.RGBA, x, y int, r, g, b, a uint8) {
	// 对点周围的像素进行强度计算
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			px, py := x+dx, y+dy
			if px < 0 || px >= img.Bounds().Dx() || py < 0 || py >= img.Bounds().Dy() {
				continue
			}

			// 计算距离和强度
			dist := math.Sqrt(float64(dx*dx + dy*dy))
			if dist > 1.0 {
				continue
			}
			intensity := 1.0 - dist

			// 混合颜色
			existing := img.RGBAAt(px, py)
			newR := uint8(float64(existing.R)*intensity + float64(r)*(1.0-intensity))
			newG := uint8(float64(existing.G)*intensity + float64(g)*(1.0-intensity))
			newB := uint8(float64(existing.B)*intensity + float64(b)*(1.0-intensity))
			newA := uint8(float64(existing.A)*intensity + float64(a)*(1.0-intensity))

			img.Set(px, py, color.RGBA{newR, newG, newB, newA})
		}
	}
}

// 三次贝塞尔曲线的参数方程
func cubic(p0, p1, p2, p3, t float64) float64 {
	return math.Pow(1-t, 3)*p0 +
		3*math.Pow(1-t, 2)*t*p1 +
		3*(1-t)*t*t*p2 +
		math.Pow(t, 3)*p3
}

// 简单的填充算法
func floodFill(img *image.RGBA, x, y int, c color.Color) {
	bounds := img.Bounds()
	target := img.At(x, y)
	if colorEquals(target, c) {
		return
	}

	queue := [][2]int{{x, y}}

	for len(queue) > 0 {
		p := queue[0]
		queue = queue[1:]

		if !colorEquals(img.At(p[0], p[1]), target) {
			continue
		}

		img.Set(p[0], p[1], c)

		// 检查四个方向
		directions := [][2]int{{0, 1}, {1, 0}, {0, -1}, {-1, 0}}
		for _, d := range directions {
			newX, newY := p[0]+d[0], p[1]+d[1]
			if newX >= bounds.Min.X && newX < bounds.Max.X &&
				newY >= bounds.Min.Y && newY < bounds.Max.Y &&
				colorEquals(img.At(newX, newY), target) {
				queue = append(queue, [2]int{newX, newY})
			}
		}
	}
}

// 比较两个颜色是否相等
func colorEquals(c1, c2 color.Color) bool {
	r1, g1, b1, a1 := c1.RGBA()
	r2, g2, b2, a2 := c2.RGBA()
	return r1 == r2 && g1 == g2 && b1 == b2 && a1 == a2
}

func drawSector(img *image.RGBA, centerX, centerY, radius int, startAngle, endAngle float64, fanShapeColor [3]float64) {
	// 创建超采样缓冲区
	ssaaImg := image.NewRGBA(image.Rect(0, 0, img.Bounds().Dx()*ssaaScale, img.Bounds().Dy()*ssaaScale))

	// 计算超采样后的参数
	sCenterX := centerX * ssaaScale
	sCenterY := centerY * ssaaScale
	sRadius := radius * ssaaScale

	start := startAngle * math.Pi / 180.0
	end := endAngle * math.Pi / 180.0

	// 转换颜色
	c := color.RGBA{
		R: uint8(fanShapeColor[0] * 255),
		G: uint8(fanShapeColor[1] * 255),
		B: uint8(fanShapeColor[2] * 255),
		A: 255,
	}

	// 在超采样缓冲区中绘制扇形
	for y := -sRadius - 1; y <= sRadius+1; y++ {
		for x := -sRadius - 1; x <= sRadius+1; x++ {
			// 计算点到圆心的距离
			dist := math.Sqrt(float64(x*x + y*y))

			if dist <= float64(sRadius) {
				angle := math.Atan2(float64(y), float64(x))
				if angle < 0 {
					angle += 2 * math.Pi
				}

				if isAngleBetween(angle, start, end) {
					// 计算抗锯齿的 alpha 值
					alpha := 1.0
					if dist > float64(sRadius-1) {
						alpha = float64(sRadius) - dist
					}

					// 设置像素颜色，考虑 alpha 混合
					ssaaImg.Set(sCenterX+x, sCenterY+y, color.RGBA{
						R: uint8(float64(c.R) * alpha),
						G: uint8(float64(c.G) * alpha),
						B: uint8(float64(c.B) * alpha),
						A: uint8(float64(c.A) * alpha),
					})
				}
			}
		}
	}

	// 将超采样缓冲区缩放回原始大小
	downscale(ssaaImg, img)
}

func isAngleBetween(angle, start, end float64) bool {
	if start <= end {
		return angle >= start && angle <= end
	}
	return angle >= start || angle <= end
}

func drawText(img *image.RGBA, score float64, x, y int) error {
	// 解析字体
	f, err := opentype.Parse(goregular.TTF)
	if err != nil {
		return err
	}

	// 创建字体face
	face, err := opentype.NewFace(f, &opentype.FaceOptions{
		Size:    45,
		DPI:     72,
		Hinting: font.HintingFull,
	})
	if err != nil {
		return err
	}
	defer face.Close()

	// 创建drawer
	d := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(color.Black),
		Face: face,
	}

	// 格式化文本
	text := fmt.Sprintf("%.1f", score)

	// 计算文本尺寸
	bounds, _ := d.BoundString(text)
	textWidth := bounds.Max.X - bounds.Min.X
	textHeight := bounds.Max.Y - bounds.Min.Y

	// 计算绘制位置（居中）
	px := fixed.I(x) - textWidth/2
	py := fixed.I(y) + textHeight/2

	// 设置绘制位置
	d.Dot = fixed.Point26_6{
		X: px,
		Y: py,
	}

	// 绘制文本
	d.DrawString(text)

	return nil
}

// 添加缩放函数
func downscale(ssaa *image.RGBA, dst *image.RGBA) {
	bounds := dst.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			var r, g, b, a float64

			// 采样超采样缓冲区的对应像素
			for dy := 0; dy < ssaaScale; dy++ {
				for dx := 0; dx < ssaaScale; dx++ {
					sx := x*ssaaScale + dx
					sy := y*ssaaScale + dy
					c := ssaa.RGBAAt(sx, sy)
					r += float64(c.R)
					g += float64(c.G)
					b += float64(c.B)
					a += float64(c.A)
				}
			}

			// 计算平均值
			scale := float64(ssaaScale * ssaaScale)
			dst.Set(x, y, color.RGBA{
				uint8(r / scale),
				uint8(g / scale),
				uint8(b / scale),
				uint8(a / scale),
			})
		}
	}
}
