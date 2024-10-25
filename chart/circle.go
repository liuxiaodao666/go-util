package chart

import (
	"fmt"
	"math"
	"os"
)

// 定义一个函数，用于生成一个圆形的SVG图像
func circleSvg(score float64) error {

	var largeArcFlag = "0"
	var fanShapeColor string

	angel := 360 * (score / 100)
	rad := (angel - 90) * math.Pi / 180

	yx := 150 + 90*math.Cos(rad)
	yy := 100 + 90*math.Sin(rad)

	if score > 50.0 {
		largeArcFlag = "1"
	}

	// 根据分数的大小，设置扇形的颜色
	switch {
	case score > 90.9999:
		fanShapeColor = "#b4de66"
	case score > 80.9999:
		fanShapeColor = "#f5c70f"
	case score > 60.9999:
		fanShapeColor = "#f59c3d"
	default:
		fanShapeColor = "#f76d6a"
	}

	create, err := os.Create("circle.svg")
	if err != nil {
		return err
	}
	defer create.Close()

	// 如果分数大于99.99，则将x坐标设置为149.99，因为起始点重合，扇形会消失
	if score > 99.99 {
		yx = 149.99
	}

	_, err = create.WriteString(fmt.Sprintf(circle, largeArcFlag, yx, yy, fanShapeColor, score))

	return err
}
