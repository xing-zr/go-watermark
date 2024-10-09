# go-watermark

## Install

```
go get -u github.com/xing-zr/go-watermark
```

## Usage

### Image Watermark

```go
config := ImageWatermarkConfig{
    OriginImagePath:    "./testdata/origin.jpg",
    WatermarkImagePath: "./testdata/watermark.png",
    WatermarkPos:       LeftTop,
    CompositeImagePath: "./testdata/composite.jpg",
}
CreateImageWatermark(config)
```

### Text Watermark

```go
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
CreateTextWatermark(config)
```