FROM golang:1.19 as build-helm-generate

WORKDIR /src

ADD go.mod /src/
ADD go.sum /src/
RUN go mod download

ADD . /src/
RUN CGO_ENABLED=0 GOOS=linux make all

FROM alpine:3.16.0
COPY --from=build-helm-generate /src/build/helm-generate /usr/local/bin

CMD [ "helm-generate", "." ]
