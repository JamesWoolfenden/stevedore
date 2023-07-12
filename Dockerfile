FROM alpine

RUN apk --no-cache add build-base git curl jq bash
RUN curl -s -k https://api.github.com/repos/JamesWoolfenden/stevedore/releases/latest | jq '.assets[] | select(.name | contains("linux_386")) | select(.content_type | contains("gzip")) | .browser_download_url' -r | awk '{print "curl -L -k " $0 " -o ./stevedore.tar.gz"}' | sh
RUN tar -xf ./stevedore.tar.gz -C /usr/bin/ && rm ./stevedore.tar.gz && chmod +x /usr/bin/stevedore && echo 'alias stevedore="/usr/bin/stevedore"' >> ~/.bashrc
COPY entrypoint.sh /entrypoint.sh

# Code file to execute when the docker container starts up (`entrypoint.sh`)
ENTRYPOINT ["/entrypoint.sh"]
