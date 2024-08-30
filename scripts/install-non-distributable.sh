#!/usr/bin/env bash

set -eux

INSTALL_NON_DISTRIBUTABLE=${INSTALL_NON_DISTRIBUTABLE:-}
if ! [[ ${INSTALL_NON_DISTRIBUTABLE,,} =~ ^(1|on|true|y|yes)$ ]]; then
	echo not installing non-distributable files >&2
	exit 0
fi

set -v +x

if [[ -z ${AWS_ACCESS_KEY_ID:-} ]]; then
	echo AWS_ACCESS_KEY_ID env var is missing >&2
	exit 1
fi

if [[ -z ${AWS_SECRET_ACCESS_KEY:-} ]]; then
	echo AWS_SECRET_ACCESS_KEY env var is missing >&2
	exit 1
fi

set +v -x

if [[ -z ${BUCKET:-} ]]; then
	echo BUCKET env var is missing >&2
	exit 1
fi

# a few checks before proceeding
if [[ $(basename "${PWD}") != non-distributable ]]; then
	echo "expected to be executed from within non-distributable directory"
	exit 1
fi

ARCH=$(uname -m)
if [[ $ARCH != "x86_64" ]]; then
	echo "nothing to be done for arch: $ARCH"
	exit 0
fi

# install s5cmd to fetch firmware tooling artifacts
s5cmd_url=https://github.com/peak/s5cmd/releases/download/v2.2.2/s5cmd_2.2.2_
case $(uname -s):$ARCH in
Darwin:arm64) s5cmd_url+=macOS-arm64 ;;
Darwin:x86_64) s5cmd_url+=macOS-64bit ;;
Linux:aarch64) s5cmd_url+=Linux-arm64 ;;
Linux:x86_64) s5cmd_url+=Linux-64bit ;;
esac
s5cmd_url+=.tar.gz
curl --fail --location --silent "$s5cmd_url" | tar -xz s5cmd
chmod +x s5cmd

# fetch vendor tools
cat <<-EOF | ./s5cmd run
	cp s3://${BUCKET}/IPMICFG_1.32.0_build.200910.zip .
	cp s3://${BUCKET}/SW_Broadcom_Unified_StorCLI_v007.1316.0000.0000_20200428.ZIP .
	cp s3://${BUCKET}/asrr/BIOSControl_v1.0.3.zip .
	cp s3://${BUCKET}/asrr/asrr_bios_kernel_module/asrdev-5.4.0-73-generic.ko .
	cp s3://${BUCKET}/mlxup .
	cp s3://${BUCKET}/msecli_Linux.run .
	cp s3://${BUCKET}/mvcli-4.1.13.31_A01.zip .
	cp s3://${BUCKET}/sum_2.10.0_Linux_x86_64_20221209.tar.gz .
EOF

# install dependencies

# install Mellanox mlxup
install -m 755 -D mlxup /usr/sbin/

# install SMC sum 2.10.0
tar -xvzf sum_2.10.0_Linux_x86_64_20221209.tar.gz
install -m 755 -D sum_2.10.0_Linux_x86_64/sum /usr/sbin/

# install SMC ipmicfg
unzip IPMICFG_1.32.0_build.200910.zip
install -m 755 -D IPMICFG_1.32.0_build.200910/Linux/64bit/IPMICFG-Linux.x86_64 /usr/sbin/smc-ipmicfg

# install Broadcom storcli
unzip SW_Broadcom_Unified_StorCLI_v007.1316.0000.0000_20200428.ZIP
rpm --import Unified_storcli_all_os/Linux/pubKey.asc
rpm -ivh Unified_storcli_all_os/Linux/storcli-007.1316.0000.0000-1.noarch.rpm

# install AsRockRack BIOSControl and copy kernel module
unzip BIOSControl_v1.0.3.zip
install -m 755 -D BIOSControl /usr/sbin/asrr-bioscontrol
cp asrdev*.*.ko /opt/asrr/

# install Marvell mvcli
unzip mvcli-4.1.13.31_A01.zip
install -m 755 -D mvcli-4.1.13.31_A01/x64/cli/mvcli /usr/sbin/mvcli
install -m 755 -D mvcli-4.1.13.31_A01/x64/cli/libmvraid.so /usr/lib/libmvraid.so

# install Dell msecli
chmod 755 msecli_Linux.run
./msecli_Linux.run --mode unattended

# Add symlink to location where OSIE expects vendor tools to be
ln -s /usr/sbin/sum /opt/supermicro/sum/sum
ln -s /usr/sbin/asrr-bioscontrol /usr/sbin/BIOSControl

# run ldconfig to update ld.so cache
ldconfig
