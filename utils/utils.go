package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image/png"
	"io"
	"log"
	"mime/multipart"

	"github.com/nfnt/resize"
)

// uploadFile uploads an image file to be used as a watermark for a QR code
func UploadFile(file multipart.File) ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	if _, err := io.Copy(buf, file); err != nil {
		return nil, fmt.Errorf("could not upload file. %v", err)
	}

	return buf.Bytes(), nil
}

func BuildErrorResponse(message string) []byte {
	responseData := make(map[string]string)
	responseData["error"] = message

	response, err := json.Marshal(responseData)
	if err != nil {
		log.Fatalln("Could not generate error message.")
	}

	return response
}

// resizeWatermark resizes a watermark image to the desired width and height
func ResizeWatermark(watermark io.Reader, width uint) ([]byte, error) {
	decodedImage, err := png.Decode(watermark)
	if err != nil {
		return nil, fmt.Errorf("could not decode watermark image: %v", err)
	}

	m := resize.Resize(width, 0, decodedImage, resize.Lanczos3)
	resized := bytes.NewBuffer(nil)
	png.Encode(resized, m)

	return resized.Bytes(), nil
}
