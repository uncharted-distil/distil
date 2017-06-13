FROM alpine:3.5 

RUN mkdir /distil

WORKDIR /distil

COPY distil .
COPY dist ./dist

EXPOSE 8080

ENTRYPOINT ./distil
