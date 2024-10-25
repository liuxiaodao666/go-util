package chart

var (
	circle = `
<svg width="300" height="200" xmlns="http://www.w3.org/2000/svg">
    <circle cx="150" cy="100" r="90" fill="#eeeeee" />

    <!-- 扇形 -->
    <path d="M 150 100
           L 150 10
           A 90 90 0 %s 1 %.2f %.2f
           z"
          fill="%s" />


    <circle cx="150" cy="100" r="60" fill="white" />
    <text x="150" y="115" font-size="50" text-anchor="middle" fill="#4d4d4d">%v</text>
</svg>
`
)
