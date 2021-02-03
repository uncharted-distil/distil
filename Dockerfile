FROM registry.gitlab.com/datadrivendiscovery/images/primitives:ubuntu-bionic-python36-stable-20201201-223410

# add GDAL for geospatial support
RUN sudo apt-get update && apt-get install -y \
    libgdal-dev \
    git \
    build-essential 

RUN mkdir /distil

WORKDIR /distil

COPY distil .
COPY dist ./dist
ENV PATH="${PATH}:/distil"

EXPOSE 8080

CMD distil
