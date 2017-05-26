FROM alpine:3.5 

RUN mkdir /distil-server

WORKDIR /distil-server

COPY distil-server .
COPY dist ./dist

EXPOSE 8080

ENTRYPOINT ./distil-server
