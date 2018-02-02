FROM alpine:3.5

# add bash for compatibility with NIST env
RUN apk update && apk add bash 

RUN mkdir /distil

WORKDIR /distil

COPY distil .
COPY dist ./dist
COPY deploy/data/ta3_search .
ENV PATH="${PATH}:/distil"

EXPOSE 8080

ENTRYPOINT tail -f /dev/null
