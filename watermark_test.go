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
