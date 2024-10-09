package go_watermark

import "testing"

func TestCreateImageWatermark(t *testing.T) {
	config := ImageWatermarkConfig{
		OriginImagePath:    "./testdata/origin.jpg",
		WatermarkImagePath: "./testdata/watermark.png",
		WatermarkPos:       LeftTop,
		CompositeImagePath: "./testdata/composite.jpg",
	}
	err := CreateImageWatermark(config)
	if err != nil {
		t.Error(err)
	}
}

func TestCreateTextWatermark(t *testing.T) {
	config := TextWatermarkConfig{
		OriginImagePath:    "./testdata/origin.jpg",
		CompositeImagePath: "./testdata/composite.jpg",
		FontPath:           "./testdata/font.ttf",
		TextInfos: []TextInfo{
			{
				Size: 100,
				Text: "hello world",
				X:    700,
				Y:    700,
			},
		},
	}
	err := CreateTextWatermark(config)
	if err != nil {
		t.Error(err)
	}
}
