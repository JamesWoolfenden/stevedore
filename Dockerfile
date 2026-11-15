FROM alpine:3.23.4@sha256:5b10f432ef3da1b8d4c7eb6c487f2f5a8f096bc91145e68878dd4a5019afde11

RUN apk --no-cache add build-base git curl jq bash
# VERSION is the release tag to install (e.g. v0.1.2); release.yml passes github.ref_name.
ARG VERSION
RUN set -eu; \
    : "${VERSION:?build with --build-arg VERSION=<release tag>}"; \
    base="https://github.com/JamesWoolfenden/stevedore/releases/download/${VERSION}"; \
    asset="stevedore_${VERSION#v}_linux_386.tar.gz"; \
    curl -fsSL -o "${asset}" "${base}/${asset}"; \
    curl -fsSL -o checksums.txt "${base}/checksums.txt"; \
    grep " ${asset}\$" checksums.txt | sha256sum -c -; \
    tar -xf "${asset}" -C /usr/bin/ stevedore; \
    rm "${asset}" checksums.txt; \
    chmod +x /usr/bin/stevedore; \
    echo 'alias stevedore="/usr/bin/stevedore"' >> ~/.bashrc
COPY entrypoint.sh /entrypoint.sh

# Code file to execute when the docker container starts up (`entrypoint.sh`)
ENTRYPOINT ["/entrypoint.sh"]
