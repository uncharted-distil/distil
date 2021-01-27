package util
/*
#cgo linux LDFLAGS: /usr/lib/image-upscale/image-upscale.so -L/usr/local/lib/ltensorflow -ltensorflow
#include <image-upscale/entry_functions.h>
#include <stdio.h>
float *buffer;
// assumes 4 dimensions because it is required
void createBuffer(int64_t* dimensions){
	float (*arr)[dimensions[1]][dimensions[2]][dimensions[3]] = malloc(dimensions[0] * sizeof(*arr));
	buffer = &arr[0][0][0][0];
}
int rowMajorIndex(int x, int y, int z, int ySize, int zSize){
	return z + zSize * (y + ySize * x);
}
void setBuffer(unsigned int x,unsigned int y, float r, float g,float b){
	float *p = buffer + rowMajorIndex(x,y,0,120,3);
	*p=r;
	p = buffer + rowMajorIndex(x,y,1,120,3);
	*p=g;
	p = buffer + rowMajorIndex(x,y,2,120,3);
	*p=b;
}
void printBuffer(){
	for(unsigned int i=0; i < 120; ++i)
	{
		for(unsigned int ii=0; ii < 120; ++ii)
		{
			printf("[%f,%f,%f]\n",*(buffer + rowMajorIndex(i,ii,0,120,3)),*(buffer + rowMajorIndex(i,ii,1,120,3)),*(buffer + rowMajorIndex(i,ii,2,120,3)));
			fflush(stdout);
		}
	}
}
void* getBuffer(){
	return buffer;
}
void printOutput(float* data, unsigned int numOfDimensions, int64_t* dimensions){
	for(unsigned int i=0; i < numOfDimensions; ++i)
	{
		printf("%i \n", (int)dimensions[i]);
		fflush(stdout);
	}
	for(unsigned int x=0; x < dimensions[1]; ++x)
	{
		for(unsigned int y=0; y < dimensions[2]; ++y)
		{
			int r = rowMajorIndex(x,y,0, dimensions[2], dimensions[3]);
			int g = rowMajorIndex(x,y,1, dimensions[2], dimensions[3]);
			int b = rowMajorIndex(x,y,2, dimensions[2], dimensions[3]);
			printf("[%f,%f,%f]", data[r], data[g], data[b]);
			fflush(stdout);
		}
	}
}
*/
import "C"
import (
	log "github.com/unchartedsoftware/plog"
	"github.com/pkg/errors"
	"unsafe"
	"image"
)

// LoadImageUpscaleLibrary loads the model for image upscaling
func LoadImageUpscaleLibrary(){
	buffer := make([]byte, 256)
	cStrPtr := (*C.char)(unsafe.Pointer(&buffer[0]))

	C.initialize(cStrPtr)
	if C.GoString(cStrPtr) == ""{
		log.Infof("image-upscale loaded.")
		return
	}
	log.Error(errors.New("Failed to load image-upscale"))
}
// UpscaleImage upscales the supplied image through the use machine learning
func UpscaleImage(img *image.RGBA) *image.RGBA{
	buffer := make([]byte, 256)
	cStrPtr := (*C.char)(unsafe.Pointer(&buffer[0]))
	colorDepth := 3
	imgSize := img.Bounds().Max
	dimBuffer :=[]int64{1, int64(imgSize.X), int64(imgSize.Y), int64(colorDepth)}
	dimensions := (*C.long)(unsafe.Pointer(&dimBuffer[0]))
	C.createBuffer(dimensions);

	for x := 0; x < imgSize.X; x++{
		for y :=0; y < imgSize.Y; y++{
			r,g,b,_ := img.At(x,y).RGBA() 
			fR := float64(r>>8) / 255.0
			fG := float64(g>>8) / 255.0
			fB := float64(b>>8) / 255.0
			C.setBuffer(C.uint(x),C.uint(y),C.float(fR),C.float(fG),C.float(fB));
		}
	}
	C.printBuffer()
	dataSize := C.uint(imgSize.X * imgSize.Y * colorDepth * 4)
	dataInput := C.DataInfo{numberOfDimensions:C.uint(4), dimensions:dimensions, dataType:C.TF_FLOAT, dataSize:dataSize, data:C.getBuffer()}
	output := C.runModel(cStrPtr, dataInput)
	if C.GoString(cStrPtr) != ""{
		log.Error(errors.New(C.GoString(cStrPtr)))
		// free buffer here
		return img
	}
	C.printOutput(output.buffer, output.numOfDimensions, output.dimension)
	C.freeOutputData(output)
	return img
}