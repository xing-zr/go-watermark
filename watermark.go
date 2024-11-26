package go_watermark

import (
	"errors"
	"github.com/disintegration/imaging"
	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
	"image"
	"image/color"
	"image/draw"
	"io/ioutil"
	"os"
	"path/filepath"
)

type ImageWatermarkConfig struct {
	OriginImagePath    string       // 原图地址
	WatermarkImagePath string       // 水印图地址
	WatermarkPos       watermarkPos // 水印位置
	CompositeImagePath string       // 合成图地址
	TiledRows          int          // 水印图横向平铺行数
	TiledCols          int          // 水印图横向平铺列数
}

type watermarkPos string

const (
	LeftTop     watermarkPos = "left_top"
	RightTop    watermarkPos = "right_top"
	LeftBottom  watermarkPos = "left_bottom"
	RightBottom watermarkPos = "right_bottom"
	Tiled       watermarkPos = "tiled"
)

func CreateImageWatermark(config ImageWatermarkConfig) error {
	watermarkFile, err := os.Open(config.WatermarkImagePath)
	if err != nil {
		return errors.New("open watermark image file error:" + err.Error())
	}
	defer watermarkFile.Close()

	originFile, err := os.Open(config.OriginImagePath)
	if err != nil {
		return errors.New("open origin image file error:" + err.Error())
	}
	defer originFile.Close()
	// 如果合成图片存在则删除重新生成
	isExists, _ := pathExists(config.CompositeImagePath)
	if isExists {
		err = os.Remove(config.CompositeImagePath)
		if err != nil {
			return errors.New("old composite image remove error:" + err.Error())
		}
	}
	// 判断文件夹是否存在，不存在创建
	dirPath := filepath.Dir(config.CompositeImagePath)
	isExist, _ := pathExists(dirPath)
	if !isExist {
		err = os.MkdirAll(dirPath, 0755)
		if err != nil {
			return err
		}
	}
	// 获取原图大小
	originImg, _ := imaging.Decode(originFile)
	watermarkImg, _ := imaging.Decode(watermarkFile)
	originImgWidth := originImg.Bounds().Dx()
	originImgHeight := originImg.Bounds().Dy()
	// 对水印图进行缩放(对比原图)
	targetWatermarkImgWidth := uint(originImgWidth / 5)
	destwatermarkImg := imaging.Resize(watermarkImg, int(targetWatermarkImgWidth), 0, imaging.Lanczos)

	// 根据水印位置合成图片
	var destImg image.Image
	switch config.WatermarkPos {
	case LeftTop:
		destImg = imaging.Overlay(originImg, destwatermarkImg, image.Pt(10, 10), 1)
	case RightTop:
		destImg = imaging.Overlay(originImg, destwatermarkImg, image.Pt(originImgWidth-int(targetWatermarkImgWidth)-10, 10), 1)
	case LeftBottom:
		destImg = imaging.Overlay(originImg, destwatermarkImg, image.Pt(10, originImgHeight-destwatermarkImg.Bounds().Dy()-10), 1)
	case RightBottom:
		destImg = imaging.Overlay(originImg, destwatermarkImg, image.Pt(originImgWidth-int(targetWatermarkImgWidth)-10, originImgHeight-destwatermarkImg.Bounds().Dy()-10), 1)
	case Tiled:
		if config.TiledCols == 0 || config.TiledRows == 0 {
			return errors.New("watermark position tiled need tiled_cols and tiled_rows")
		}
		mainBounds := originImg.Bounds()
		watermarkBounds := destwatermarkImg.Bounds()

		// 创建一个与主图相同尺寸的新图像作为结果图像
		result := image.NewNRGBA(mainBounds)
		draw.Draw(result, mainBounds, originImg, image.Point{}, draw.Src)

		// 计算水印在主图上平铺所需的行数和列数
		rows := config.TiledRows
		cols := config.TiledCols
		// 计算行间距和列间距
		totalWidth := cols * watermarkBounds.Dx()
		totalHeight := rows * watermarkBounds.Dy()
		extraWidth := mainBounds.Dx() - totalWidth
		extraHeight := mainBounds.Dy() - totalHeight
		rowSpacing := extraHeight / (rows + 1)
		colSpacing := extraWidth / (cols + 1)
		for r := 0; r < rows; r++ {
			for c := 0; c < cols; c++ {
				// 计算当前水印在主图上的位置
				x := c*(watermarkBounds.Dx()+colSpacing) + colSpacing/2
				y := r*(watermarkBounds.Dy()+rowSpacing) + rowSpacing/2

				// 将水印粘贴到结果图像的相应位置
				draw.DrawMask(result, image.Rect(x, y, x+watermarkBounds.Dx(), y+watermarkBounds.Dy()), destwatermarkImg, destwatermarkImg.Bounds().Min, destwatermarkImg, destwatermarkImg.Bounds().Min, draw.Over)
			}
		}
		destImg = result
	default:
		return errors.New("watermark position error")
	}
	if err = imaging.Save(destImg, config.CompositeImagePath); err != nil {
		return errors.New("create composite image error:" + err.Error())
	}
	return nil
}

type TextWatermarkConfig struct {
	OriginImagePath    string // 原图地址
	CompositeImagePath string // 合成图地址
	FontPath           string // 字体文件地址
	TextInfos          []TextInfo
}

type TextInfo struct {
	Text  string     // 文字内容
	Size  float64    // 文字大小
	Color color.RGBA // 文字颜色透明度
	X     int        // 位置信息
	Y     int        // 位置信息
}

func CreateTextWatermark(config TextWatermarkConfig) error {
	originFile, err := os.Open(config.OriginImagePath)
	if err != nil {
		return errors.New("open origin image file error:" + err.Error())
	}
	defer originFile.Close()
	// 如果合成图片存在则删除重新生成
	isExists, _ := pathExists(config.CompositeImagePath)
	if isExists {
		err = os.Remove(config.CompositeImagePath)
		if err != nil {
			return errors.New("old composite image remove error:" + err.Error())
		}
	}
	// 判断文件夹是否存在，不存在创建
	dirPath := filepath.Dir(config.CompositeImagePath)
	isExist, _ := pathExists(dirPath)
	if !isExist {
		err = os.MkdirAll(dirPath, 0755)
		if err != nil {
			return err
		}
	}
	img, err := imaging.Decode(originFile)
	if err != nil {
		return err
	}
	dst := image.NewRGBA(image.Rect(0, 0, img.Bounds().Dx(), img.Bounds().Dy()))
	draw.Draw(dst, dst.Bounds(), img, img.Bounds().Min, draw.Over)
	// load font file
	fontBytes, err := ioutil.ReadFile(config.FontPath)
	if err != nil {
		return err
	}
	font, err := freetype.ParseFont(fontBytes)
	if err != nil {
		return err
	}
	for _, v := range config.TextInfos {
		f := freetype.NewContext()
		f.SetDPI(72)
		f.SetFont(font)       // 加载字体
		f.SetFontSize(v.Size) // 设置字体尺寸
		f.SetClip(dst.Bounds())
		f.SetDst(dst)
		f.SetSrc(image.NewUniform(v.Color)) // 设置字体颜色
		// 位置信息
		pt := freetype.Pt(v.X, v.Y)
		_, err = f.DrawString(v.Text, pt)
		if err != nil {
			return err
		}
	}
	if err = imaging.Save(dst, config.CompositeImagePath); err != nil {
		return errors.New("create composite image error:" + err.Error())
	}
	return nil
}

type TextTiledWatermarkConfig struct {
	OriginImagePath    string     // 原图地址
	CompositeImagePath string     // 合成图地址
	FontPath           string     // 字体文件地址
	Text               string     // 文字内容
	Color              color.RGBA // 文字颜色透明度
	TiledRows          int        // 水印图横向平铺行数
	TiledCols          int        // 水印图横向平铺列数
}

func CreateTextTiledWatermark(config TextTiledWatermarkConfig) error {

	if config.TiledCols == 0 || config.TiledRows == 0 {
		return errors.New("watermark position tiled need tiled_cols and tiled_rows")
	}
	originFile, err := os.Open(config.OriginImagePath)
	if err != nil {
		return errors.New("open origin image file error:" + err.Error())
	}
	defer originFile.Close()
	// 如果合成图片存在则删除重新生成
	isExists, _ := pathExists(config.CompositeImagePath)
	if isExists {
		err = os.Remove(config.CompositeImagePath)
		if err != nil {
			return errors.New("old composite image remove error:" + err.Error())
		}
	}
	// 判断文件夹是否存在，不存在创建
	dirPath := filepath.Dir(config.CompositeImagePath)
	isExist, _ := pathExists(dirPath)
	if !isExist {
		err = os.MkdirAll(dirPath, 0755)
		if err != nil {
			return err
		}
	}
	watermarkImg, err := textToImage(config)
	if err != nil {
		return err
	}
	originImg, _ := imaging.Decode(originFile)
	originImgWidth := originImg.Bounds().Dx()
	//originImgHeight := originImg.Bounds().Dy()
	// 对水印图进行缩放(对比原图)
	targetWatermarkImgWidth := uint(originImgWidth / 5)
	destwatermarkImg := imaging.Resize(watermarkImg, int(targetWatermarkImgWidth), 0, imaging.Lanczos)

	mainBounds := originImg.Bounds()
	watermarkBounds := destwatermarkImg.Bounds()

	// 创建一个与主图相同尺寸的新图像作为结果图像
	result := image.NewNRGBA(mainBounds)
	draw.Draw(result, mainBounds, originImg, image.Point{}, draw.Src)

	// 计算水印在主图上平铺所需的行数和列数
	rows := config.TiledRows
	cols := config.TiledCols
	// 计算行间距和列间距
	totalWidth := cols * watermarkBounds.Dx()
	totalHeight := rows * watermarkBounds.Dy()
	extraWidth := mainBounds.Dx() - totalWidth
	extraHeight := mainBounds.Dy() - totalHeight
	rowSpacing := extraHeight / (rows + 1)
	colSpacing := extraWidth / (cols + 1)
	for r := 0; r < rows; r++ {
		for c := 0; c < cols; c++ {
			// 计算当前水印在主图上的位置
			x := c*(watermarkBounds.Dx()+colSpacing) + colSpacing/2
			y := r*(watermarkBounds.Dy()+rowSpacing) + rowSpacing/2

			// 将水印粘贴到结果图像的相应位置
			draw.DrawMask(result, image.Rect(x, y, x+watermarkBounds.Dx(), y+watermarkBounds.Dy()), destwatermarkImg, destwatermarkImg.Bounds().Min, destwatermarkImg, destwatermarkImg.Bounds().Min, draw.Over)
		}
	}
	if err = imaging.Save(result, config.CompositeImagePath); err != nil {
		return errors.New("create composite image error:" + err.Error())
	}
	return nil
}

func textToImage(config TextTiledWatermarkConfig) (*image.NRGBA, error) {
	// 加载字体文件
	fontData, err := os.ReadFile(config.FontPath)
	if err != nil {
		return nil, errors.New("failed to read font file:" + err.Error())
	}

	fontFace, err := truetype.Parse(fontData)
	if err != nil {
		return nil, errors.New("failed to parse font:" + err.Error())
	}

	// 设置字体大小
	var fontSize float64 = 50
	face := truetype.NewFace(fontFace, &truetype.Options{
		Size:    fontSize,
		DPI:     72,
		Hinting: font.HintingFull,
	})
	// 计算文字的宽度和高度
	textWidth, textHeight := measureText(face, config.Text)
	// 创建一个新的图像
	img := image.NewRGBA(image.Rect(0, 0, textWidth, textHeight))
	// 绘制背景颜色透明
	draw.Draw(img, img.Bounds(), image.Transparent, image.ZP, draw.Src)
	// 设置文字颜色并添加透明度
	//textColor := color.RGBA{R: 0, G: 0, B: 0, A: 200} // 黑色，半透明
	textColor := config.Color
	// 绘制文字
	d := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(textColor),
		Face: face,
	}
	// 设置绘制起点
	d.Dot = fixed.P(0, textHeight-5)
	// 绘制文字
	d.DrawString(config.Text)
	// 图片旋转
	dst := imaging.Rotate(img, 45, color.Transparent)

	return dst, nil
}

// measureText 计算给定文字的宽度和高度
func measureText(face font.Face, text string) (int, int) {
	var (
		width  int
		height int
	)
	// 获取字体的度量信息
	metrics := face.Metrics()

	// 遍历每个字符，计算总宽度
	for _, textRune := range text {
		// 获取字符的水平间距
		advance, _ := face.GlyphAdvance(textRune)
		width += int(advance >> 6)
	}

	// 计算高度
	height = int(metrics.Height >> 6)

	return width, height
}
