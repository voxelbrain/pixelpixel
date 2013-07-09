body {
	padding: {{.Spacing}}px;
}

#pixels {
	width: {{.TotalWidth}}px;
	margin: 0 auto;
	font-size: 0;
}

#pixels canvas {
	background-color: black;
	display: inline-block;
	width: {{.PixelSize}}px;
	height: {{.PixelSize}}px;
	margin-right: {{.Spacing}}px;
	margin-bottom: {{.Spacing}}px;
}
