FROM golang:1.19

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . ./

WORKDIR cmd/fmtconvert/
RUN CGO_ENABLED=0 GOOS=linux go build -v
RUN go install -v

WORKDIR /data
ENTRYPOINT ["fmtconvert"]
