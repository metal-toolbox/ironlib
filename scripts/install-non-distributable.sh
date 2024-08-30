#!/bin/bash
set -eux

INSTALL_NON_DISTRIBUTABLE=${INSTALL_NON_DISTRIBUTABLE:-}
if ! [[ ${INSTALL_NON_DISTRIBUTABLE,,} =~ ^(1|on|true|y|yes)$ ]]; then
	echo not installing non-distributable files >&2
	exit 0
fi

ARCH=$(uname -m)
export WORKDIR="non-distributable"

# a few checks before proceeding
PARENT_DIR="$(basename "$(pwd)")"

[[ "$PARENT_DIR" != "${WORKDIR}" ]] && echo "expected to be executed from within ${WORKDIR} directory" && exit 1
[[ "${ARCH}" != "x86_64" ]] && echo "nothing to be done for arch: ${ARCH}" && exit 0

# minio client s3 parameters
# https://github.com/minio/mc/blob/master/docs/minio-client-complete-guide.md#specify-temporary-host-configuration-through-environment-variable
export MC_HOST_${S3_BUCKET_ALIAS}="https://${ACCESS_KEY}:${SECRET_KEY}@s3.amazonaws.com"

export ASRDEV_KERNEL_MODULE=asrdev-5.4.0-73-generic.ko

# set minio client url
OS=$(uname -s)
ARCH=$(uname -m)

MC_ARCH=$ARCH

MC_ARCH=$ARCH

case $ARCH in
aarch64 | arm64)
	MC_ARCH=arm64
	;;
x86_64)
	MC_ARCH=amd64
	;;
*)
	echo "unsupported ARCH; $ARCH"
	exit 1
	;;
esac

case $OS in
Darwin*)
	MC_URL="https://dl.min.io/client/mc/release/darwin-${MC_ARCH}/mc"
	;;
Linux*)
	MC_URL="https://dl.min.io/client/mc/release/linux-${MC_ARCH}/mc"
	;;
*)
	echo "unsupported OS: $OS"
	exit 1
	;;
esac

# install minio client to fetch firmware tooling artifacts
curl "${MC_URL}" -o mc && chmod +x mc

# fetch vendor tools
./mc cp "${S3_PATH}"/mlxup .
./mc cp "${S3_PATH}"/msecli_Linux.run .
./mc cp "${S3_PATH}"/IPMICFG_1.32.0_build.200910.zip .
./mc cp "${S3_PATH}"/sum_2.10.0_Linux_x86_64_20221209.tar.gz .
./mc cp "${S3_PATH}"/SW_Broadcom_Unified_StorCLI_v007.1316.0000.0000_20200428.ZIP .
./mc cp "${S3_PATH}"/asrr/BIOSControl_v1.0.3.zip .
./mc cp "${S3_PATH}"/asrr/asrr_bios_kernel_module/$ASRDEV_KERNEL_MODULE .
./mc cp "${S3_PATH}"/mvcli-4.1.13.31_A01.zip .

# install dependencies
#
# install Mellanox mlxup
install -m 755 -D mlxup /usr/sbin/

# install SMC sum 2.10.0
tar -xvzf sum_2.10.0_Linux_x86_64_20221209.tar.gz &&
	install -m 755 -D sum_2.10.0_Linux_x86_64/sum /usr/sbin/

# install SMC ipmicfg
unzip IPMICFG_1.32.0_build.200910.zip &&
	install -m 755 -D IPMICFG_1.32.0_build.200910/Linux/64bit/IPMICFG-Linux.x86_64 /usr/sbin/smc-ipmicfg

# install Broadcom storcli
unzip SW_Broadcom_Unified_StorCLI_v007.1316.0000.0000_20200428.ZIP &&
	rpm --import Unified_storcli_all_os/Linux/pubKey.asc &&
	rpm -ivh Unified_storcli_all_os/Linux/storcli-007.1316.0000.0000-1.noarch.rpm

# install AsRockRack BIOSControl and copy kernel module
unzip BIOSControl_v1.0.3.zip &&
	install -m 755 -D BIOSControl /usr/sbin/asrr-bioscontrol &&
	cp asrdev*.*.ko /opt/asrr/

# install Marvell mvcli
unzip mvcli-4.1.13.31_A01.zip &&
	install -m 755 -D mvcli-4.1.13.31_A01/x64/cli/mvcli /usr/sbin/mvcli &&
	install -m 755 -D mvcli-4.1.13.31_A01/x64/cli/libmvraid.so /usr/lib/libmvraid.so

# install Dell msecli
chmod 755 msecli_Linux.run && ./msecli_Linux.run --mode unattended

# Add symlink to location where OSIE expects vendor tools to be
ln -s /usr/sbin/sum /opt/supermicro/sum/sum
ln -s /usr/sbin/asrr-bioscontrol /usr/sbin/BIOSControl

# run ldconfig to update ld.so cache
ldconfig
