#!/bin/bash
set -ex

test -d ${DEPDIR} && rm -rf ./"${DEPDIR}"
mkdir -p ${DEPDIR}
cd ${DEPDIR}

# download epel repo config
curl -sO https://dl.fedoraproject.org/pub/epel/epel-release-latest-9.noarch.rpm

# Go back to root of checkout
cd ..

# Provide a hook for downloading extra dependencies, so downstream users of ironlib
# can include any extra dependencies they need
if [ -f "extra-dependencies.sh" ]; then
	./extra-dependencies.sh
fi