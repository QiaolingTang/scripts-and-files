FROM registry.access.redhat.com/ubi8/ubi-minimal:latest
LABEL name="http-echo-receiver" \
      summary="A simple ncat-based HTTP request logger for OCP"
RUN microdnf install -y nmap-ncat && \
    microdnf clean all
USER 1001
EXPOSE 8080
CMD ["ncat", "-l", "-p", "8080", "-k", "-vv"] 
