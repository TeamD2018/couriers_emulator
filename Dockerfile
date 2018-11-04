FROM golang:1.11

WORKDIR /go/src

ENV PATH="/go/bin:${PATH}"

COPY . .

ENV GO111MODULE=on

RUN go install

EXPOSE 2018

CMD ["courier_emulator",  "--backend", "http://dc.utkin.xyz:8081", "--routes", "http://dc.utkin.xyz:5000", "--mode", "s", "-c", "10", "-t", "1000"]