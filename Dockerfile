FROM ubuntu:18.04

# add GDAL for geospatial support
RUN apt-get update && \
    apt-get install -y software-properties-common curl libpq-dev && \
    rm -rf /var/lib/apt/lists/* 
RUN curl -sL https://deb.nodesource.com/setup_10.x | bash - && apt-get install -y nodejs
RUN add-apt-repository ppa:ubuntugis/ppa && \
apt-get update && \
apt-get -y install build-essential openssh-client git wget unzip gdal-bin gdal-data libgdal-dev libgdal-perl libgdal-perl-doc python3-gdal


RUN wget -O go.tar.gz https://golang.org/dl/go1.15.7.linux-amd64.tar.gz && tar -C /usr/local -xzf go.tar.gz
# setup golang env vars
ENV PATH="/usr/local/go/bin:$PATH"
ENV GOPATH=/opt/go
ENV PATH=$PATH:$GOPATH/bin
# gdal env vars
ENV CPLUS_INCLUDE_PATH=/usr/include/gdal
ENV C_INCLUDE_PATH=/usr/include/gdal
RUN npm install -g yarn

RUN mkdir /distil

WORKDIR /distil

COPY distil .
COPY dist ./dist
ENV PATH="${PATH}:/distil"

EXPOSE 8080

CMD distil
