# Use the full RHEL 9 Universal Base Image
FROM registry.access.redhat.com/ubi9/ubi:latest

LABEL name="http-echo-receiver-rhel9" \
      summary="RHEL 9 ncat-based HTTP request logger"

# 'dnf' is available here by default. 
# We install nmap-ncat (which provides 'ncat').
RUN dnf install -y nmap-ncat && \
    dnf clean all

# OCP requires a non-root user for the 'restricted' SCC.
# UBI 9 images are designed to run as any UID, but 1001 is the convention.
USER 1001

EXPOSE 8080

# -l: listen, -p: port, -k: keep open for multiple requests, -vv: very verbose
CMD ["ncat", "-l", "-p", "8080", "-k", "-vv"]