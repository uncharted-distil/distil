tensorflow_dir="/usr/local/tensorflow"
mac_tensorflow_tar="libtensorflow-cpu-darwin-x86_64-2.4.0.tar.gz"
linux_tensorflow_cpu_tar="libtensorflow-cpu-linux-x86_64-2.4.0.tar.gz"
linux_tensorflow_gpu_tar="libtensorflow-gpu-linux-x86_64-2.4.0.tar.gz"
tensorflow_url_mac="https://storage.googleapis.com/tensorflow/libtensorflow/libtensorflow-cpu-darwin-x86_64-2.4.0.tar.gz"
tensorflow_url_linux_cpu="https://storage.googleapis.com/tensorflow/libtensorflow/libtensorflow-cpu-linux-x86_64-2.4.0.tar.gz"
tensorflow_url_linux_gpu="https://storage.googleapis.com/tensorflow/libtensorflow/libtensorflow-gpu-linux-x86_64-2.4.0.tar.gz"
image_upscale_lib="/usr/local/image-upscale.so"
image_upscale_url_linux_cpu="https://github.com/uncharted-distil/distil-image-upscale/releases/download/1.0-linux-cpu/image-upscale.so"
image_upscale_url_linux_gpu="https://github.com/uncharted-distil/distil-image-upscale/releases/download/1.0-linux-gpu/image-upscale.so"
image_upscale_url_mac="https://github.com/uncharted-distil/distil-image-upscale/releases/download/1.0-mac-cpu/image-upscale.so"
image_upscale_src="https://github.com/uncharted-distil/distil-image-upscale/archive/1.0-linux-gpu.tar.gz"
image_upscale_src_tar="1.0-linux-gpu.tar.gz"
image_src_dir="distil-image-upscale-1.0-linux-gpu/src"
uname=`uname`
models_dir="./static_resources/models"
user_lib_dir="/usr/local"
source_dir="/usr/include/image-upscale"
usr_include_dir="/usr/include"
if [ "$uname" = Darwin ]; then
    usr_include_dir="/usr/local/include"
    source_dir="/usr/local/include/image-upscale"
fi
if ! command -v wget &> /dev/null;then
    echo "missing required tool wget please install"
    exit 1
fi
get_tensorflow(){
    local tensorflow_dir="/usr/local/tensorflow"
    echo $tensorflow_dir
    mkdir $tensorflow_dir
    wget $2 -P $tensorflow_dir
    tar -C $tensorflow_dir -xzf $tensorflow_dir/$1
}
# check if tensorflow lib is installed
if [ ! -d "$tensorflow_dir" ]; then
    echo "unable to locate tensorflow lib"
    if [ "$uname" = Linux ]; then
        # if it fails cuda is not installed so get the tensorflow cpu
        if command -v nvcc &> /dev/null; then
            echo "cuda not found fetching tensorflow cpu"
            get_tensorflow "$linux_tensorflow_cpu_tar" "$tensorflow_url_linux_cpu"
        else
            # cuda exists get tensorflow for gpu
            echo "cuda found fetching tensorflow gpu"
            get_tensorflow $linux_tensorflow_gpu_tar $tensorflow_url_linux_gpu
        fi 
    fi
    if [ "$uname" = 'Darwin' ]; then
        echo "fetching mac tensorflow binaries"
        # get mac binaries for tensorflow c
        get_tensorflow $mac_tensorflow_tar $tensorflow_url_mac
    fi
fi
# check if image_upscale.so is installed
if [ ! -f "$image_upscale_lib" ]; then
    if [ "$uname" = Linux ]; then
        if command -v nvcc &> /dev/null; then
            echo "cuda not found fetching image-upscale cpu"
            wget $image_upscale_url_linux_cpu -P $user_lib_dir
        else
            echo "cuda found fetching image-upscale gpu"
            wget $image_upscale_url_linux_gpu -P $user_lib_dir
        fi 
    fi
    if [ "$uname" = 'Darwin' ]; then
        echo "fetching image-upscale"
        wget $image_upscale_url_mac -P $user_lib_dir
    fi
fi
# check if source is installed
if [ ! -d "$source_dir" ]; then
    echo "fetching image-scale source"
    wget $image_upscale_src -P $user_lib_dir
    # extract
    tar -C $user_lib_dir -xvzf $user_lib_dir/$image_upscale_src_tar
    mkdir $source_dir
    # copy source over
    cp -a $user_lib_dir/$image_src_dir/. $source_dir
fi
# check if models are in static folder
if [ ! -d "$models_dir" ]; then
    echo "unable to locate model weights"
    echo "fetching model weights"
    # should fetch model weights from somewhere
fi