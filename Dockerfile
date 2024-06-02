# build ironlib wrapper binaries
FROM golang:1.22-alpine AS stage0

WORKDIR /workspace

# copy the go modules manifests
COPY go.mod go.mod
COPY go.sum go.sum

# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# copy ironlib go sources
COPY . .

# build helper util
ARG TARGETOS TARGETARCH
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
    go build -o getbiosconfig examples/biosconfig/biosconfig.go && \
    install -m 755 -D getbiosconfig /usr/sbin/

RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
    go build -o getinventory examples/inventory/inventory.go && \
    install -m 755 -D getinventory /usr/sbin/

FROM almalinux:9-minimal as stage1
ARG TARGETOS TARGETARCH

# copy ironlib wrapper binaries
COPY --from=stage0 /usr/sbin/getbiosconfig /usr/sbin/getbiosconfig
COPY --from=stage0 /usr/sbin/getinventory /usr/sbin/getinventory

# import and install tools
RUN curl -sO https://dl.fedoraproject.org/pub/epel/epel-release-latest-9.noarch.rpm
RUN microdnf install -y --setopt=tsflags=nodocs crypto-policies-scripts && \
    update-crypto-policies --set DEFAULT:SHA1 && \
    rpm -ivh epel-release-latest-9.noarch.rpm && \
    rm -f epel-release-latest-9.noarch.rpm

## Prerequisite directories for Dell, ASRR, Supermicro
## /lib/firmware required for Dell updates to be installed successfullly
RUN mkdir -p /lib/firmware /opt/asrr /usr/libexec/dell_dup /opt/supermicro/sum/

RUN echo "Target ARCH is $TARGETARCH"

# Bootstrap Dell DSU repository
# Install Dell idracadm7 to enable collecting BIOS configuration and use install_weak_deps=0 to avoid pulling extra packages
RUN if [[ $TARGETARCH = "amd64" ]] ; then \
    curl -O https://linux.dell.com/repo/hardware/dsu/bootstrap.cgi && \
    bash bootstrap.cgi && rm -f bootstrap.cgi && \
    microdnf install -y --setopt=tsflags=nodocs --setopt=install_weak_deps=0 \
    srvadmin-idracadm7 ; fi

# update dependencies
RUN microdnf update -y --setopt=tsflags=nodocs --setopt=install_weak_deps=0 \
                       --setopt=keepcache=0 && microdnf clean all

# install misc support packages
RUN microdnf install -y --setopt=tsflags=nodocs --setopt=install_weak_deps=0 \
    dmidecode     \
    dosfstools    \
    e2fsprogs     \
    gdisk         \
    gzip          \
    hdparm        \
    kmod          \
    libssh2-devel \
    lshw          \
    mdadm         \
    nvme-cli      \
    pciutils      \
    smartmontools \
    tar           \
    udev          \
    unzip         \
    util-linux    \
    python        \
    python-devel  \
    python-pip  \
    python-setuptools \
    which        \
    mstflint &&  \
    microdnf clean all && \
    ln -s /usr/bin/microdnf /usr/bin/yum

RUN pip install uefi_firmware==v1.11

# Install our custom flashrom package
ADD https://github.com/metal-toolbox/flashrom/releases/download/v1.3.99/flashrom-1.3.99-0.el9.x86_64.rpm /tmp
RUN if [[ $TARGETARCH = "amd64" ]] ; then \
    rpm -ivh /tmp/flashrom*.rpm ; fi

# Delete /tmp/* as we don't need those included in the image.
RUN rm -rf /tmp/*

# Install non-distributable files when the env var is set.
#
# The non-distributable files are executables provided by hardware vendors.
ARG INSTALL_NON_DISTRIBUTABLE=false
ENV INSTALL_NON_DISTRIBUTABLE=$INSTALL_NON_DISTRIBUTABLE

# S3_BUCKET_ALIAS is the alias set on the S3 bucket, for details refer to the minio
# client guide https://github.com/minio/mc/blob/master/docs/minio-client-complete-guide.md
ARG S3_BUCKET_ALIAS=utils
ENV S3_BUCKET_ALIAS=$S3_BUCKET_ALIAS

# S3_PATH is the path in the s3 bucket where the non-distributable files are located
# note, this generally includes the s3 bucket alias
ARG S3_PATH
ENV S3_PATH=$S3_PATH

ARG ACCESS_KEY
ENV ACCESS_KEY=$ACCESS_KEY

ARG SECRET_KEY
ENV SECRET_KEY=$SECRET_KEY

COPY scripts scripts
RUN if [[ $INSTALL_NON_DISTRIBUTABLE = "true" ]]; then \
    mkdir -p non-distributable && \
    cp scripts/install-non-distributable.sh ./non-distributable/install.sh && \
    cd ./non-distributable/ && \
    ./install.sh $S3_BUCKET_ALIAS && \
    cd .. && rm -rf non-distributable; fi
RUN rm -rf scripts/

# Build a lean image with dependencies installed.
FROM scratch
COPY --from=stage1 / /

ENTRYPOINT [ "/bin/bash", "-l", "-c" ]
