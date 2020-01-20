FROM golang
ADD . /go/voting-app
ENV GIN_MODE release

RUN cd /go/voting-app; go mod download; go build -o /go/voting .; cp -r  /go/voting-app/templates /go/

# Run the outyet command by default when the container starts.
ENTRYPOINT ["/go/voting"]

# Expose the server TCP port
EXPOSE 8080