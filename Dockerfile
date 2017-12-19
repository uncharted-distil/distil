FROM alpine:3.5

RUN mkdir /distil

WORKDIR /distil

COPY distil .
COPY dist ./dist
COPY deploy/data/ta3_search .

EXPOSE 8080

ENTRYPOINT ./distil
