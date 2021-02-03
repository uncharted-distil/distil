tensorflow_dir="/usr/local/tensorflow"
mac_tensorflow_tar="libtensorflow-cpu-darwin-x86_64-2.4.0.tar.gz"
linux_tensorflow_cpu_tar="libtensorflow-cpu-linux-x86_64-2.4.0.tar.gz"
linux_tensorflow_gpu_tar="libtensorflow-gpu-linux-x86_64-2.4.0.tar.gz"
tensorflow_url_mac="https://storage.googleapis.com/tensorflow/libtensorflow/libtensorflow-cpu-darwin-x86_64-2.4.0.tar.gz"
tensorflow_url_linux_cpu="https://storage.googleapis.com/tensorflow/libtensorflow/libtensorflow-cpu-linux-x86_64-2.4.0.tar.gz"
tensorflow_url_linux_gpu="https://storage.googleapis.com/tensorflow/libtensorflow/libtensorflow-gpu-linux-x86_64-2.4.0.tar.gz"
image_upscale_lib="/usr/local/image-upscale.so"
image_upscale_url="https://github.com/uncharted-distil/distil-image-upscale/archive/master.zip"
image_upscale_src_zip="master.zip"
src_folder="distil-image-upscale-master"
image_src_dir="$src_folder/src"
image_model_dir="$src_folder/models"
uname=`uname`
static_resources="./static_resources"
models_dir="$static_resources/models"
local_dir="/usr/local"
source_dir="/usr/include/image-upscale"
usr_include_dir="/usr/include"
if [ "$uname" = Darwin ]; then
    usr_include_dir="/usr/local/include"
    source_dir="/usr/local/include/image-upscale"
fi
if ! command -v wget > /dev/null 2>&1;then
    echo "missing required tool wget please install"
    exit 1
fi
if ! command -v unzip > /dev/null 2>&1; then
    echo "missing required tool unzip please install"
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
        if ! [ -x "$(command -v nvcc)" ]; then
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
rm -rf "$source_dir" || true
echo "fetching image-scale source"
wget $image_upscale_url -P $local_dir
# extract
unzip $local_dir/$image_upscale_src_zip -d $local_dir
mkdir $source_dir
# copy source over
cp -a $local_dir/$image_src_dir/. $source_dir
# check if models are in static folder
if [ ! -d "$models_dir" ]; then
    echo "unable to locate model weights"
    echo "fetching model weights"
    mkdir -p "$static_resources/models"
    # should fetch model weights from somewhere
    cp -a $local_dir/$image_model_dir/ "$static_resources/models"
fi

echo "cleaning up files"
# cleanup
rm -rf $local_dir/$src_folder || true
rm -rf $local_dir/$image_upscale_src_zip || true
