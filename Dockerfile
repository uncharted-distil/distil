FROM ubuntu:18.04

# add GDAL for geospatial support
RUN apt-get update && \
    apt-get install -y software-properties-common curl libpq-dev && \
    rm -rf /var/lib/apt/lists/*
RUN curl -sL https://deb.nodesource.com/setup_10.x | bash - && apt-get install -y nodejs
RUN add-apt-repository ppa:ubuntugis/ppa && \
    apt-get update && \
    apt-get -y install build-essential openssh-client git unzip wget gdal-bin gdal-data libgdal-dev

# add tensorflow
RUN mkdir /usr/local/tensorflow && \
    wget https://storage.googleapis.com/tensorflow/libtensorflow/libtensorflow-gpu-linux-x86_64-2.4.0.tar.gz -P /usr/local/tensorflow && \
    tar -C /usr/local/tensorflow -xzf /usr/local/tensorflow/libtensorflow-gpu-linux-x86_64-2.4.0.tar.gz && \
    mkdir -p /usr/local/lib && \
    cp -a /usr/local/tensorflow/lib/. /usr/local/lib  && \
    rm -rf /usr/local/tensorflow && \
    ldconfig

RUN mkdir /distil

WORKDIR /distil

COPY distil .
COPY dist ./dist
COPY static_resources ./static_resources

ENV PATH="${PATH}:/distil"

EXPOSE 8080

CMD distil
