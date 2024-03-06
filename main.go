package main

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"net/http"
	"strconv"

	"github.com/Abhi-singh-karuna/GO-QR/utils"
	"github.com/gin-gonic/gin"
	qrcode "github.com/skip2/go-qrcode"
)

const MAX_UPLOAD_SIZE = 1024 * 1024 

type simpleQRCode struct {
	Content string
	Size    int
}

func (code *simpleQRCode) Generate() ([]byte, error) {
	qrCode, err := qrcode.Encode(code.Content, qrcode.Medium, code.Size)
	if err != nil {
		return nil, fmt.Errorf("could not generate a QR code: %v", err)
	}
	return qrCode, nil
}

func (code *simpleQRCode) GenerateWithWatermark(watermark []byte) ([]byte, error) {
	qrCode, err := code.Generate()
	if err != nil {
		return nil, err
	}

	qrCode, err = code.addWatermark(qrCode, watermark, code.Size)
	if err != nil {
		return nil, fmt.Errorf("could not add watermark to QR code: %v", err)
	}

	return qrCode, nil
}

// addWatermark adds a watermark to a QR code, centered in the middle of the QR code
func (code *simpleQRCode) addWatermark(qrCode []byte, watermarkData []byte, size int) ([]byte, error) {
	qrCodeData, err := png.Decode(bytes.NewBuffer(qrCode))
	if err != nil {
		return nil, fmt.Errorf("could not decode QR code: %v", err)
	}

	watermarkImage, err := png.Decode(bytes.NewBuffer(watermarkData))
	if err != nil {
		return nil, fmt.Errorf("could not decode watermark: %v", err)
	}

	// Determine the offset to center the watermark on the QR code
	offset := image.Pt(((size / 2) - 32), ((size / 2) - 32))

	watermarkImageBounds := qrCodeData.Bounds()
	m := image.NewRGBA(watermarkImageBounds)

	// Center the watermark over the QR code
	draw.Draw(m, watermarkImageBounds, qrCodeData, image.Point{}, draw.Src)
	draw.Draw(
		m,
		watermarkImage.Bounds().Add(offset),
		watermarkImage,
		image.Point{},
		draw.Over,
	)

	watermarkedQRCode := bytes.NewBuffer(nil)
	png.Encode(watermarkedQRCode, m)

	return watermarkedQRCode.Bytes(), nil
}

func main() {
	app := gin.Default()

	app.GET("/ping", pingFunction)
	app.POST("/generate", handleRequest)

	app.Run(":8080")
}

func pingFunction(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "pong",
	})
}

func handleRequest(c *gin.Context) {

	// read form-data of request

	size, url := c.Request.FormValue("size"), c.Request.FormValue("url")

	fmt.Println(size, url)

	// request.ParseMultipartForm(10 << 20)
	// var size, url string = request.FormValue("size"), request.FormValue("url")
	// var codeData []byte

	// if url == "" {
	// 	writer.Write(buildErrorResponse("Could not determine the desired QR code content."))
	// 	writer.WriteHeader(400)
	// 	return
	// }

	qrCodeSize, err := strconv.Atoi(size)
	if err != nil || size == "" {
		c.JSON(http.StatusBadRequest, gin.H{"Could not determine the desired QR code size :": err.Error()})
		return
	}

	qrCode := simpleQRCode{Content: url, Size: qrCodeSize}
	var codeData []byte

	watermarkFile, _, err := c.Request.FormFile("watermark")
	if err != nil && errors.Is(err, http.ErrMissingFile) {
		fmt.Println("Watermark image was not uploaded or could not be retrieved. Reason: ", err)
		codeData, err = qrCode.Generate()
		if err != nil {

			c.JSON(http.StatusBadRequest, gin.H{"Could not generate QR code.": err.Error()})
			return
		}

		c.Writer.Header().Set("Content-Type", "image/png")
		c.Writer.Write(codeData)
		return
	}

	watermark, err := utils.UploadFile(watermarkFile)
	if err != nil {

		c.JSON(http.StatusBadRequest, gin.H{"Could not upload the watermark image.": err.Error()})
		return
	}

	contentType := http.DetectContentType(watermark)
	if err != nil {
		_ = utils.BuildErrorResponse(fmt.Sprintf("Provided watermark image is a %s not a PNG. %v.", err, contentType))
		c.JSON(http.StatusBadRequest, gin.H{"Provided watermark image is a child ,,, .": err.Error()})
		return
	}

	// waterMarkSize, err := strconv.Atoi(size)
	// if err != nil {
	// 	c.JSON(http.StatusBadRequest, gin.H{"Error during converting integer from string": err.Error()})
	// 	return
	// }

	waterMarkSize, err := strconv.ParseUint(size, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error during converting unit from string": err.Error()})
	}
	x := uint(waterMarkSize)

	watermark, err = utils.ResizeWatermark(bytes.NewBuffer(watermark), x/4)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Could not resize the watermark image.": err.Error()})
		return
	}

	codeData, err = qrCode.GenerateWithWatermark(watermark)
	if err != nil {
		_ = utils.BuildErrorResponse(fmt.Sprintf("Could not generate QR code with the watermark image. %v", err))
		return
	}

	c.Writer.Header().Set("Content-Type", "image/png")
	c.Writer.Write(codeData)

	c.JSON(http.StatusOK, qrCode)
}
