# Use the full RHEL 9 Universal Base Image
FROM registry.access.redhat.com/ubi9/ubi:latest

LABEL name="http-echo-receiver-rhel9" \
      summary="RHEL 9 ncat-based HTTP request logger"

RUN dnf install -y nmap-ncat && \
    dnf clean all

USER 1001

EXPOSE 8080

# -l: Listen, -p: Port, -k: Keep open, -vv: Very Verbose
CMD ["ncat", "-l", "-p", "8080", "-k", "-vv"]
