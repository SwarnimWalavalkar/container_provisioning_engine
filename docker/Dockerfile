FROM debian:bookworm

RUN apt-get -y update; apt-get -y install curl

COPY . .

# Install Docker Container Engine
RUN curl -LO https://download.docker.com/linux/debian/dists/bookworm/pool/stable/amd64/docker-ce-cli_24.0.7-1~debian.12~bookworm_amd64.deb
RUN dpkg -i docker-ce-cli_24.0.7-1~debian.12~bookworm_amd64.deb

# Install Docker Buildx
RUN curl -LO https://download.docker.com/linux/debian/dists/bookworm/pool/stable/amd64/docker-buildx-plugin_0.11.2-1~debian.12~bookworm_amd64.deb
RUN dpkg -i docker-buildx-plugin_0.11.2-1~debian.12~bookworm_amd64.deb

# Install Go
RUN curl -LO https://dl.google.com/go/go1.20.linux-amd64.tar.gz
RUN tar -C /usr/local -xzf go1.20.linux-amd64.tar.gz

ENV PATH="$PATH:/usr/local/go/bin"

RUN go mod download
RUN go build

CMD [ "./container_provisioning_engine" ]
