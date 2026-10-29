FROM scratch
COPY validkit /validkit
EXPOSE 7700
ENTRYPOINT ["/validkit", "-no-auth"]
