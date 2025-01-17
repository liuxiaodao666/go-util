package png

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"strconv"
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
	fanShapeEndAngle = math.Floor(360*(scoreFloat/100) + 270)
	fmt.Println(fanShapeEndAngle)
	//扇形颜色
	fanShapeColor = getFanShapeColor(callBackMedia.Assessment)
	//fmt.Printf("Fan shape color: %s\n", fanShapeColor)

	drawPng(fanShapeColor, fanShapeEndAngle, scoreFloat, dirPath)

	return dirPath, nil
}

func drawPng(fanShapeColor [3]float64, fanShapeEndAngle, scoreFloat float64, pngName string) {
	// 创建一个新的 RGBA 图像
	img := image.NewRGBA(image.Rect(0, 0, 300, 200))

	// 填充背景为白色
	// draw.Draw(img, img.Bounds(), &image.Uniform{color.White}, image.Point{}, draw.Src)

	// 绘制灰色背景圆
	drawCircle(img, 150, 100, 90, color.RGBA{237, 237, 237, 255})
	tempSave(img, "1.png")
	// 绘制扇形
	// 当结束角度超过360度时，需要分两次绘制
	if fanShapeEndAngle > 360 {
		// 先绘制270到360度的部分
		drawSector(img, 150, 100, 90, 270, 360, fanShapeColor)
		// 再绘制0到剩余角度的部分
		drawSector(img, 150, 100, 90, 0, fanShapeEndAngle-360, fanShapeColor)
	} else {
		drawSector(img, 150, 100, 90, 270, fanShapeEndAngle, fanShapeColor)
	}
	tempSave(img, "2.png")
	// 绘制中心白色圆
	drawCircle(img, 150, 100, 50, color.White)

	// 绘制文字
	err := drawText(img, scoreFloat, 150, 100)
	if err != nil {
		fmt.Println(err)
	}

	// 保存为PNG文件
	f, _ := os.Create(pngName)
	defer f.Close()
	png.Encode(f, img)
}

func tempSave(img image.Image, pngName string) {
	f, _ := os.Create(pngName)
	defer f.Close()
	png.Encode(f, img)
}

// 添加抗锯齿相关常量
const (
	kappa      = 0.5522847498
	ssaaScale  = 4      // 超采样倍数
	bezierStep = 0.0001 // 贝塞尔曲线采样精度
)

// 绘制带抗锯齿的圆形
func drawCircle(img *image.RGBA, centerX, centerY, radius int, c color.Color) {
	ssaaImg := image.NewRGBA(image.Rect(0, 0, img.Bounds().Dx()*ssaaScale, img.Bounds().Dy()*ssaaScale))

	// 计算超采样后的参数
	sCenterX := centerX * ssaaScale
	sCenterY := centerY * ssaaScale
	sRadius := radius * ssaaScale
	offset := int(float64(sRadius) * kappa)

	// 定义圆的基准点
	points := []struct{ x, y int }{
		{sCenterX, sCenterY - sRadius},
		{sCenterX + sRadius, sCenterY},
		{sCenterX, sCenterY + sRadius},
		{sCenterX - sRadius, sCenterY},
	}

	// 定义控制点
	controls := []struct{ x1, y1, x2, y2 int }{
		{
			sCenterX + offset, sCenterY - sRadius,
			sCenterX + sRadius, sCenterY - offset,
		},
		{
			sCenterX + sRadius, sCenterY + offset,
			sCenterX + offset, sCenterY + sRadius,
		},
		{
			sCenterX - offset, sCenterY + sRadius,
			sCenterX - sRadius, sCenterY + offset,
		},
		{
			sCenterX - sRadius, sCenterY - offset,
			sCenterX - offset, sCenterY - sRadius,
		},
	}

	// 绘制贝塞尔曲线到超采样缓冲区
	for i := 0; i < 4; i++ {
		p0 := points[i]
		p1 := controls[i]
		p2 := points[(i+1)%4]

		drawAntialiasedBezier(ssaaImg,
			p0.x, p0.y,
			p1.x1, p1.y1,
			p1.x2, p1.y2,
			p2.x, p2.y,
			c)
	}

	// 填充圆形内部
	floodFill(ssaaImg, sCenterX, sCenterY, c)
	// 将超采样缓冲区缩放回原始大小并进行平滑处理
	for y := 0; y < img.Bounds().Dy(); y++ {
		for x := 0; x < img.Bounds().Dx(); x++ {
			existingColor := img.RGBAAt(x, y)
			sx := x * ssaaScale
			sy := y * ssaaScale

			var r, g, b, a float64
			samples := float64(ssaaScale * ssaaScale)
			targetR, targetG, targetB, targetA := c.RGBA()

			// 转换目标颜色到0-255范围
			targetR = targetR >> 8
			targetG = targetG >> 8
			targetB = targetB >> 8
			targetA = targetA >> 8

			// 收集超采样像素
			for dy := 0; dy < ssaaScale; dy++ {
				for dx := 0; dx < ssaaScale; dx++ {
					ssaaColor := ssaaImg.RGBAAt(sx+dx, sy+dy)
					if ssaaColor.A > 0 {
						// 使用目标颜色的RGB值，只从采样中获取alpha
						r += float64(targetR) / samples
						g += float64(targetG) / samples
						b += float64(targetB) / samples
						a += float64(ssaaColor.A) / samples
					}
				}
			}

			// 如果新颜色的alpha值足够大，则完全使用目标颜色
			if a > 250 {
				img.Set(x, y, c)
			} else if a > 0 {
				// alpha混合
				alpha := a / 255.0
				newR := uint8(float64(existingColor.R)*(1-alpha) + float64(targetR)*alpha)
				newG := uint8(float64(existingColor.G)*(1-alpha) + float64(targetG)*alpha)
				newB := uint8(float64(existingColor.B)*(1-alpha) + float64(targetB)*alpha)
				newA := uint8(math.Max(float64(existingColor.A), a))

				img.Set(x, y, color.RGBA{newR, newG, newB, newA})
			}
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
	ssaaImg := image.NewRGBA(image.Rect(0, 0, img.Bounds().Dx()*ssaaScale, img.Bounds().Dy()*ssaaScale))

	sCenterX := centerX * ssaaScale
	sCenterY := centerY * ssaaScale
	sRadius := (radius - 1) * ssaaScale

	start := startAngle * math.Pi / 180.0
	end := endAngle * math.Pi / 180.0

	// 预乘alpha的颜色值
	c := color.RGBA{
		R: uint8(fanShapeColor[0] * 255),
		G: uint8(fanShapeColor[1] * 255),
		B: uint8(fanShapeColor[2] * 255),
		A: 255,
	}

	// 绘制扇形到超采样缓冲区
	for y := -sRadius - 1; y <= sRadius+1; y++ {
		for x := -sRadius - 1; x <= sRadius+1; x++ {
			dist := math.Sqrt(float64(x*x + y*y))

			// 计算边缘平滑度
			edgeDist := math.Abs(dist - float64(sRadius))
			if dist > float64(sRadius+1) {
				continue
			}

			angle := math.Atan2(float64(y), float64(x))
			if angle < 0 {
				angle += 2 * math.Pi
			}

			if isAngleBetween(angle, start, end) {
				// 计算alpha值，考虑边缘平滑
				alpha := float64(1.0)
				if edgeDist < 1.0 {
					alpha = 1.0 - edgeDist
				}

				// 使用预乘alpha的颜色值
				col := color.RGBA{
					R: uint8(float64(c.R) * alpha),
					G: uint8(float64(c.G) * alpha),
					B: uint8(float64(c.B) * alpha),
					A: uint8(255 * alpha),
				}
				ssaaImg.Set(sCenterX+x, sCenterY+y, col)
			}
		}
	}

	// 将超采样缓冲区缩放回原始大小
	for y := 0; y < img.Bounds().Dy(); y++ {
		for x := 0; x < img.Bounds().Dx(); x++ {
			existingColor := img.RGBAAt(x, y)
			sx := x * ssaaScale
			sy := y * ssaaScale

			var r, g, b float64
			var totalAlpha float64
			samples := float64(ssaaScale * ssaaScale)

			// 收集并平均超采样像素
			for dy := 0; dy < ssaaScale; dy++ {
				for dx := 0; dx < ssaaScale; dx++ {
					ssaaColor := ssaaImg.RGBAAt(sx+dx, sy+dy)
					alpha := float64(ssaaColor.A) / 255.0
					r += float64(ssaaColor.R) / samples
					g += float64(ssaaColor.G) / samples
					b += float64(ssaaColor.B) / samples
					totalAlpha += alpha / samples
				}
			}

			if totalAlpha > 0 {
				// 使用预乘alpha进行颜色混合
				finalAlpha := totalAlpha
				if finalAlpha > 1.0 {
					finalAlpha = 1.0
				}

				newR := uint8(r + float64(existingColor.R)*(1.0-finalAlpha))
				newG := uint8(g + float64(existingColor.G)*(1.0-finalAlpha))
				newB := uint8(b + float64(existingColor.B)*(1.0-finalAlpha))

				img.Set(x, y, color.RGBA{newR, newG, newB, 255})
			}
		}
	}
}

func isAngleBetween(angle, start, end float64) bool {
	if start <= end {
		return angle >= start && angle <= end
	}
	return angle >= start || angle <= end
}

// drawText 在指定位置绘制居中的文本
func drawText(img *image.RGBA, score float64, centerX, centerY int) error {
	font := NewFont("numbers", 32, 64)
	font.SetCharSpacing(1)
	font.SetScale(0.8)

	// 将分数转换为字符串
	text := fmt.Sprintf("%v", score)

	// 计算缩放后的字符尺寸
	charWidth, charHeight := font.GetScaledSize()
	scaledSpacing := int(float64(font.CharSpacing) * font.Scale)

	// 计算整个文本的总宽度
	totalWidth := len(text)*(charWidth+scaledSpacing) - scaledSpacing // 减去最后一个字符后的间距

	// 计算起始位置，使文本居中
	startX := centerX - totalWidth/2
	startY := centerY - charHeight/2

	// 绘制文本
	err := font.DrawText(img, startX, startY, text, color.RGBA{0, 0, 0, 255})
	if err != nil {
		return err
	}

	return nil
}
