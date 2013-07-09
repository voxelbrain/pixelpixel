body {
	padding: {{.Spacing}}px;
}

#pixels {
	width: {{.TotalWidth}}px;
	margin: 0 auto;
	font-size: 0;
}

.pixel {
	background-color: black;
	display: inline-block;
	width: {{.PixelSize}}px;
	height: {{.PixelSize}}px;
	margin-right: {{.Spacing}}px;
	margin-bottom: {{.Spacing}}px;
	border: 5px solid white;
}

.pixel.error {
	border: 5px solid red;
}
