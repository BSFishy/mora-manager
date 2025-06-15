FROM golang:1.24 as build

WORKDIR /app

COPY . .

RUN go install github.com/a-h/templ/cmd/templ@v0.3.865 \
  && templ generate \
  && CGO_ENABLED=0 GOOS=linux go build -o runway .

FROM gcr.io/distroless/static-debian12@sha256:b7b9a6953e7bed6baaf37329331051d7bdc1b99c885f6dbeb72d75b1baad54f9

COPY --from=build /app/runway /

CMD [ "/runway" ]
