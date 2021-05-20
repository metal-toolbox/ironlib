FROM centos AS stage0
ARG FUP_FILES_SOURCE=http://install.packet.net/firmware/fup

## collect vendor tooling artifacts
RUN dnf install -y make unzip
## fetch vendor tools
## TODO: switch these to a public s3 bucket
RUN curl -sO $FUP_FILES_SOURCE/image-tooling/mlxup && \
    curl -sO $FUP_FILES_SOURCE/image-tooling/msecli_Linux.run && \
    curl -sO $FUP_FILES_SOURCE/image-tooling/IPMICFG_1.32.0_build.200910.zip && \
    curl -sO $FUP_FILES_SOURCE/image-tooling/sum_2.5.0_Linux_x86_64_20200722.tar.gz && \
    curl -sO $FUP_FILES_SOURCE/image-tooling/SW_Broadcom_Unified_StorCLI_v007.1316.0000.0000_20200428.ZIP && \
    # install mlxup
    install -m 755 -D mlxup /usr/sbin/ && \
    # install SMC sum 2.5.0
    tar -xvzf sum_2.5.0_Linux_x86_64_20200722.tar.gz && \
    install -m 755 -D sum_2.5.0_Linux_x86_64/sum /usr/sbin/ && \
    # install SMC ipmicfg
    unzip IPMICFG_1.32.0_build.200910.zip && \
    install -m 755 -D IPMICFG_1.32.0_build.200910/Linux/64bit/IPMICFG-Linux.x86_64 /usr/sbin/smc-ipmicfg && \
    # install storecli
    unzip SW_Broadcom_Unified_StorCLI_v007.1316.0000.0000_20200428.ZIP


# build ironlib wrapper binaries
FROM golang:1.16-alpine AS stage1

WORKDIR /workspace

# copy the go modules manifests
COPY go.mod go.mod
COPY go.sum go.sum

# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# copy the go sources
COPY . .

# build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -o getbiosconfig examples/biosconfig/biosconfig.go && \
    install -m 755 -D getbiosconfig /usr/sbin/


# main
FROM centos
# copy vendor tooling artifacts
COPY --from=stage0 /usr/sbin/mlxup /usr/sbin/mlxup
COPY --from=stage0 /usr/sbin/sum /usr/sbin/sum
COPY --from=stage0 /usr/sbin/smc-ipmicfg /usr/sbin/smc-ipmicfg
COPY --from=stage0 /fup/Unified_storcli_all_os/Linux/pubKey.asc /tmp/storecli_pubkey.asc
COPY --from=stage0 /fup/Unified_storcli_all_os/Linux/storcli-007.1316.0000.0000-1.noarch.rpm /tmp/
COPY --from=stage0 /fup/msecli_Linux.run /tmp/
# copy ironlib wrapper binaries
COPY --from=stage1 /usr/sbin/getbiosconfig /usr/sbin/getbiosconfig
# import and install tools
RUN rpm --import /tmp/storecli_pubkey.asc && \
    dnf install -y /tmp/storcli-007.1316.0000.0000-1.noarch.rpm && \
    chmod 755 /tmp/msecli_Linux.run && /tmp/msecli_Linux.run --mode unattended

############# Dell ####################
COPY dell-system-update.repo /etc/yum.repos.d/
## Dell BIOS updates fail if this folder doesn't exist
RUN mkdir -p /lib/firmware

# Add keys required by the dell-system-update utility
RUN mkdir -p /usr/libexec/dell_dup && cd  /usr/libexec/dell_dup && \
    curl -sO https://linux.dell.com/repo/pgp_pubkeys/0x756ba70b1019ced6.asc && \
    curl -sO https://linux.dell.com/repo/pgp_pubkeys/0xca77951d23b66a9d.asc && \
    curl -sO https://linux.dell.com/repo/pgp_pubkeys/0x1285491434D8786F.asc && \
    curl -sO https://linux.dell.com/repo/pgp_pubkeys/0x3CA66B4946770C59.asc

# install misc support packages
RUN dnf install -y https://dl.fedoraproject.org/pub/epel/epel-release-latest-8.noarch.rpm && \
    dnf install -y vim \
                   tar  \
                   lshw  \
                   unzip  \
                   nano    \
                   gzip     \
                   less      \
                   which      \
                   strace      \
                   pciutils     \
                   passwd        \
                   nvme-cli       \
                   dmidecode       \
                   libssh2-devel    \
                   smartmontools     \
                   'dnf-command(config-manager)' && \
    dnf config-manager --disable production-dell-system-update_dependent && \
    dnf config-manager --disable production-dell-system-update_independent

ENTRYPOINT [ "/bin/bash", "-l", "-c" ]
