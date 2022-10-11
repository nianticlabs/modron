FROM --platform=linux/amd64 golang:1.18-alpine

RUN apk -U add git gcompat tzdata zip ca-certificates terraform python3 npm
# Install the Google Cloud SDK. We could use the google image, but alpine there is severely outdated.
# Downloading gcloud package
RUN wget https://dl.google.com/dl/cloudsdk/release/google-cloud-sdk.tar.gz -O /tmp/google-cloud-sdk.tar.gz

# Installing the package
RUN mkdir -p /usr/local/gcloud \
    && tar -C /usr/local/gcloud -xvf /tmp/google-cloud-sdk.tar.gz \
    && /usr/local/gcloud/google-cloud-sdk/install.sh

# Adding the package path to local
ENV PATH $PATH:/usr/local/gcloud/google-cloud-sdk/bin

# Install protoc release 3.20.1
# Downloading protoc package
# TODO: Upgrade to 21.x once issue https://github.com/protocolbuffers/protobuf-javascript/issues/127 is fixed
RUN wget https://github.com/protocolbuffers/protobuf/releases/download/v3.20.1/protoc-3.20.1-linux-x86_64.zip -O /tmp/protoc-3.20.1-linux-x86_64.zip

# Installing the protoc package
RUN mkdir -p /usr/local/protoc
RUN unzip /tmp/protoc-3.20.1-linux-x86_64.zip -d /usr/local/protoc

# Adding the package path to local
ENV PATH $PATH:/usr/local/protoc/bin

# Copy sources over and install dependencies.
ENV CGO_ENABLED=0

ADD . /workspaces/modron
WORKDIR /workspaces/modron/src
RUN git config --global --add safe.directory /workspaces/modron

# Install protoc-gen-go and -grpc plugins for protoc
# For some reason, protoc does not find the plugin even though
# /go/bin is added to PATH. Add `--plugin=/go/bin/protoc-gen-go` to
# the protoc invocation to fix this for now.
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
RUN go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

RUN npm install --global ts-protoc-gen
RUN npm install --global @angular/cli

ENV PATH $PATH:/usr/local/lib/node_modules/ts-protoc-gen/bin/
RUN go mod download

# Persist bashrc.
RUN SNIPPET="export PROMPT_COMMAND='history -a' && export HISTFILE=/commandhistory/.bash_history" && echo $SNIPPET >> "/root/.bashrc"
