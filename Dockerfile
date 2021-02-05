FROM ubuntu:18.04

# add GDAL for geospatial support
RUN apt-get update && \
    apt-get install -y software-properties-common curl libpq-dev && \
    rm -rf /var/lib/apt/lists/* 
RUN curl -sL https://deb.nodesource.com/setup_10.x | bash - && apt-get install -y nodejs
RUN add-apt-repository ppa:ubuntugis/ppa && \
apt-get update && \
apt-get -y install build-essential openssh-client git wget gdal-bin gdal-data libgdal-dev libgdal-perl libgdal-perl-doc python3-gdal

RUN mkdir /distil

WORKDIR /distil

COPY distil .
COPY dist ./dist
# copy tensorflow libs
COPY /usr/local/tensorflow/lib /usr/local/lib
# copy tensorflow include
COPY /usr/local/tensorflow/include /usr/local/include
# copy image-upscale source over
COPY /usr/local/include/image-upscale /usr/local/include/image-upscale
ENV PATH="${PATH}:/distil"

EXPOSE 8080

CMD distil
