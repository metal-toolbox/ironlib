# build ironlib wrapper binaries
FROM golang:1.18-alpine AS stage0

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
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on \
    go build -o getbiosconfig examples/biosconfig/biosconfig.go && \
    install -m 755 -D getbiosconfig /usr/sbin/

FROM almalinux:9-minimal as stage1

ARG DEPDIR="dependencies"
COPY "${DEPDIR}" dependencies

# copy ironlib wrapper binaries
COPY --from=stage0 /usr/sbin/getbiosconfig /usr/sbin/getbiosconfig

# import and install tools
RUN microdnf install -y --setopt=tsflags=nodocs crypto-policies-scripts && \
    update-crypto-policies --set DEFAULT:SHA1 && \
    rpm -ivh ${DEPDIR}/epel-release-latest-9.noarch.rpm

## Prerequisite directories for Dell, ASRR, Supermicro
## /lib/firmware required for Dell updates to be installed successfullly
RUN mkdir -p /lib/firmware /opt/asrr /usr/libexec/dell_dup /opt/supermicro/sum/
# Bootstrap Dell DSU repository
RUN curl -O https://linux.dell.com/repo/hardware/dsu/bootstrap.cgi && bash bootstrap.cgi && rm -f bootstrap.cgi
# Install Dell idracadm7 to enable collecting BIOS configuration and use install_weak_deps=0 to avoid pulling extra packages
RUN microdnf install -y --setopt=tsflags=nodocs --setopt=install_weak_deps=0 \
    srvadmin-idracadm7

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
    unzip         \
    util-linux    \
    which &&      \
    microdnf clean all && \
    ln -s /usr/bin/microdnf /usr/bin/yum # since dell dsu expects yum

# Provide hook to include extra dependencies in the image
RUN if [[ -f ${DEPDIR}/install-extra-deps.sh ]]; then cd ${DEPDIR} && bash install-extra-deps.sh; fi

# Delete /tmp/* and dependencies dir as we don't need those included in the image.
RUN rm -rf /tmp/* && rm -rf ${DEPDIR}

# Build a lean image with dependencies installed.
FROM scratch
COPY --from=stage1 / /

ENTRYPOINT [ "/bin/bash", "-l", "-c" ]
