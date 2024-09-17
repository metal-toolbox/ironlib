# build ironlib wrapper binaries
FROM golang:1.22-alpine AS helper-binaries

WORKDIR /workspace

# copy the go modules manifests
COPY go.mod go.sum ./

# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# copy rest of go sources
COPY . .

# build helper util
ENV CGO_ENABLED=0
RUN go build -o /usr/sbin/getbiosconfig examples/biosconfig/biosconfig.go
RUN go build -o /usr/sbin/getinventory examples/inventory/inventory.go

FROM almalinux:9-minimal AS base
FROM base AS deps

# Configure microdnf to avoid installing unwanted packages
RUN printf 'install_weak_deps=0\ntsflags=nodocs\n' >>/etc/dnf/dnf.conf

# import and install tools
RUN curl -sO https://dl.fedoraproject.org/pub/epel/epel-release-latest-9.noarch.rpm
RUN microdnf install -y crypto-policies-scripts && \
    update-crypto-policies --set DEFAULT:SHA1 && \
    rpm -ivh epel-release-latest-9.noarch.rpm && \
    rm -f epel-release-latest-9.noarch.rpm

## Prerequisite directories for Dell, ASRR, Supermicro
## /lib/firmware required for Dell updates to be installed successfullly
RUN mkdir -p /lib/firmware /opt/asrr /opt/supermicro/sum/ /usr/libexec/dell_dup

ARG TARGETARCH
RUN echo "Target ARCH is $TARGETARCH"

# Bootstrap Dell DSU repository
# Install Dell idracadm7 to enable collecting BIOS configuration and use install_weak_deps=0 to avoid pulling extra packages
RUN if [[ $TARGETARCH == "amd64" ]]; then \
      curl -O https://linux.dell.com/repo/hardware/dsu/bootstrap.cgi && \
      bash bootstrap.cgi && \
      rm -f bootstrap.cgi && \
      microdnf install -y srvadmin-idracadm7; \
    fi

# update dependencies
RUN microdnf update -y && microdnf clean all

# install misc support packages
RUN microdnf install -y \
      dmidecode         \
      dosfstools        \
      e2fsprogs         \
      gdisk             \
      gzip              \
      hdparm            \
      kmod              \
      libssh2-devel     \
      lshw              \
      mdadm             \
      nvme-cli          \
      pciutils          \
      python            \
      python-devel      \
      python-pip        \
      python-setuptools \
      smartmontools     \
      tar               \
      udev              \
      unzip             \
      util-linux        \
      which             \
      && \
    microdnf clean all && \
    ln -s /usr/bin/microdnf /usr/bin/yum

RUN pip install uefi_firmware==v1.11

# Install our custom flashrom package
ADD https://github.com/metal-toolbox/flashrom/releases/download/v1.3.99/flashrom-1.3.99-0.el9.x86_64.rpm /tmp
RUN if [[ $TARGETARCH == "amd64" ]] ; then \
      rpm -ivh /tmp/flashrom*.rpm; \
    fi

# Delete /tmp/* as we don't need those included in the image.
RUN rm -rf /tmp/*

# Use same base image used by deps so that we keep all the meta-vars (CMD, PATH, ...)
FROM base
# copy ironlib wrapper binaries
COPY --from=helper-binaries /usr/sbin/getbiosconfig /usr/sbin/getinventory /usr/sbin/
# copy installed packages
COPY --from=deps / /

# Install non-distributable files when the env var is set, left for downstream consumers
COPY scripts/install-non-distributable.sh non-distributable/
ONBUILD ARG AWS_ACCESS_KEY_ID
ONBUILD ARG AWS_SECRET_ACCESS_KEY
ONBUILD ARG BUCKET
ONBUILD ARG INSTALL_NON_DISTRIBUTABLE
ONBUILD RUN cd non-distributable && ./install-non-distributable.sh
ONBUILD RUN rm -rf non-distributable
