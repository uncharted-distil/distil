FROM docker-hub.uncharted.software/alpine:3.11.6

# add bash + packages to support CGO
RUN apk update && apk add bash git make build-base

# add GDAL for geospatial support
RUN apk add gdal gdal-dev

RUN mkdir /distil

WORKDIR /distil

COPY distil .
COPY dist ./dist
ENV PATH="${PATH}:/distil"

EXPOSE 8080

CMD distil
