package util
/*
#cgo linux LDFLAGS: -ldl

#include <dlfcn.h>
#include <limits.h>
#include <stdint.h>
#include <stdlib.h>
#include <stdio.h>

void* initialize=NULL;
void* runModel=NULL;
void* freeOutputData=NULL;
void* cleanup=NULL;

static void* functionLookup(void* handle, const char* functionName, char** error){
	void* func = dlsym(handle, functionName);
	if(!func){
		*error=(char*)dlerror();
		return NULL;
	}
	return func;
}
static unsigned int loadPlugin(const char* dllPath, char** error){
	void* handle = dlopen(dllPath, RTLD_NOW | RTLD_GLOBAL);
	if(!handle){
		*error=(char*)dlerror();
		return 0;
	}
	initialize=functionLookup(handle, "initialize", error);
	if(!initialize){
		return 0;
	}
	runModel=functionLookup(handle, "runModel", error);
	if(!runModel){
		return 0;
	}
	freeOutputData=functionLookup(handle, "freeOutputData", error);
	if(!freeOutputData){
		return 0;
	}
	cleanup=functionLookup(handle, "cleanup", error);
	if(!cleanup){
		return 0;
	}
	return 1;
}
*/
import "C"
const (

)
func loadImageUpscaleLibrary(){
	C.loadPlugin(C.CString("image-upscale.so"),)
}