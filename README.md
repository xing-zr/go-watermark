# go-watermark

## Install

```
go get -u github.com/xing-zr/go-watermark
```

## Usage

```go
config := ImageWatermarkConfig{
    OriginImagePath:    "./testdata/origin.jpg",
    WatermarkImagePath: "./testdata/watermark.png",
    WatermarkPos:       LeftTop,
    CompositeImagePath: "./testdata/composite.jpg",
}
CreateImageWatermark(config)
```