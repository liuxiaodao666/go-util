package png

import (
	"fmt"
	"image"
	"image/color"

	"math"
	"strconv"

	"github.com/llgcode/draw2d"
	"github.com/llgcode/draw2d/draw2dimg"
)

//配置字体路径
// func init() {
// draw2d.SetFontFolder("../resource/font")
// }

/* Started by AICoder, pid:ha78fuad26t156e140cc0817c10a511b5a7420c4 */
// getFanShapeColor 根据评估等级返回颜色
func getFanShapeColor(assessment string) color.RGBA {
	colorMap := map[string]color.RGBA{
		"excellent": {180, 222, 102, 255}, //"#b4de66"
		"good":      {245, 199, 15, 255},  //"#f5c70f"
		"normal":    {245, 156, 61, 255},  //"#f59c3d"
		"poor":      {247, 109, 106, 255}, //"#f76d6a"
	}
	return colorMap[assessment]
}

type CallBackMedia struct {
	HealthScore string `json:"health_score"`
	Assessment  string `json:"assessment"`
}

// createSVGFile 创建 SVG 文件并返回文件路径
func CreateSVGFile(dirPath string, callBackMedia CallBackMedia) (string, error) {
	var fanShapeColor color.RGBA
	var scoreFloat float64
	var fanShapeAngle float64
	// 将字符串转换为浮点数
	scoreFloat, err := strconv.ParseFloat(callBackMedia.HealthScore, 64)
	if err != nil {
		return "", fmt.Errorf("Error converting score to float: %v", err)
	}
	//fmt.Printf("Parsed score as float: %.1f\n", scoreFloat)
	//计算扇形角度
	fanShapeAngle = 360 * (scoreFloat / 100)
	//扇形颜色
	fanShapeColor = getFanShapeColor(callBackMedia.Assessment)
	//fmt.Printf("Fan shape color: %s\n", fanShapeColor)
	drawPng(fanShapeColor, fanShapeAngle, scoreFloat, dirPath)
	return dirPath, nil
}
func drawPng(fanShapeColor color.RGBA, fanShapeAngle, scoreFloat float64, pngName string) {
	// 创建画布
	dest := image.NewRGBA(image.Rect(0, 0, 300, 200))
	gc := draw2dimg.NewGraphicContext(dest)
	// 绘制大圆
	gc.SetFillColor(color.RGBA{0xee, 0xee, 0xee, 0xff})
	gc.SetStrokeColor(color.Transparent) // 设置边线为透明
	// 或者直接设置线宽为0
	gc.SetLineWidth(0)
	gc.BeginPath()
	gc.ArcTo(150, 100, 90, 90, 0, 2*math.Pi)
	gc.Fill()
	// // 绘制扇形
	drawSector(gc, 150, 100, 90, 270, fanShapeAngle, fanShapeColor)
	// 绘制小圆（白色）
	gc.SetFillColor(color.White)
	gc.BeginPath()
	gc.ArcTo(150, 100, 50, 50, 0, 2*math.Pi)
	gc.Fill()
	// 添加文字
	gc.SetFontData(draw2d.FontData{Name: "luxi", Family: draw2d.FontFamilySans, Style: draw2d.FontStyleNormal})
	gc.SetFontSize(35)
	gc.SetFillColor(color.Black)
	// 计算文本位置以居中显示
	text := fmt.Sprintf("%v", scoreFloat)

	// 获取文本度量信息
	left, top, right, bottom := gc.GetStringBounds(text)
	textWidth := right - left
	textHeight := bottom - top

	// 计算居中位置
	x := 150 - textWidth/2  // 150 是画布宽度的一半
	y := 100 + textHeight/2 // 100 是画布高度的一半

	// 绘制文本
	gc.FillStringAt(text, x, y)
	// 保存图片
	draw2dimg.SaveToPngFile(pngName, dest)
}
func drawSector(gc *draw2dimg.GraphicContext, x, y, radius, startAngle, angle float64, fanShapeColor color.RGBA) {
	// 将角度转换为弧度
	start := startAngle * math.Pi / 180.0
	changeAngle := angle * math.Pi / 180.0
	gc.SetFillColor(fanShapeColor)
	gc.SetLineWidth(0.1)
	gc.BeginPath()
	// 移动到圆心
	gc.MoveTo(x, y)
	// 画一条线到扇形的起点
	gc.LineTo(x+radius*math.Cos(start), y+radius*math.Sin(start))
	// 画弧线
	gc.ArcTo(x, y, radius, radius, start, changeAngle)
	// 闭合路径
	gc.Close()
	gc.Fill()
}
