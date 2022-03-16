# Based on AWS provided images for Lambda latest version of x86_64 arch.
# STAGE 1: Build
FROM public.ecr.aws/lambda/provided:latest-x86_64 AS build

# Configure Go
ENV GOROOT /usr/local/go
ENV GOPATH /workspace/go
ENV GOBIN $GOPATH/bin
ENV PATH $PATH:$GOROOT/bin:$GOBIN

RUN mkdir -p $GOROOT $GOPATH $GOBIN

# Install Go 1.17
WORKDIR /tmp/
RUN yum -y install tar gzip make
RUN curl https://dl.google.com/go/go1.17.7.linux-amd64.tar.gz -o go1.17.7.linux-amd64.tar.gz
RUN rm -rf $GOROOT && tar -C /usr/local -xzf ./go1.17.7.linux-amd64.tar.gz
RUN rm -rf go1.17.7.linux-amd64.tar.gz

# Build app
WORKDIR /functions/
COPY ./ ./
RUN make GO_ARCH="amd64"

# STAGE 2: Deployment
# Keep base image the same with STAGE 1
FROM public.ecr.aws/lambda/provided:latest-x86_64

# Install functions
COPY --from=build /functions/bin/ ${LAMBDA_TASK_ROOT}
# Install RIE script
COPY --from=build /functions/entry_script.sh /entry_script.sh
#
ENTRYPOINT [ "/entry_script.sh" ]
