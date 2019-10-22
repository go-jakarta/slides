FROM golang:1.13

ADD ./webapp/webapp /usr/bin/

EXPOSE 8080

WORKDIR /root

CMD ["/usr/bin/webapp", "-l=0.0.0.0:8080", "-db=postgres://postgres:P4ssw0rd@postgres"]
