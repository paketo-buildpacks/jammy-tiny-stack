#!/bin/bash

# TODO: Why do we install packages this way instead of apt-get installing them?
set -eu -o pipefail

mkdir -p "/tiny/var/lib/dpkg/status.d"
apt download $(cat packagelist | tr '\n' ' ')

while read PACKAGE; do
    echo "installing $PACKAGE..."
    # TODO: Maybe don't unpack right into the build root?
    ar x $PACKAGE*.deb
    ls -al *
    if [[ "$(ls data.*)" == "data.tar.xz" ]]; then
      ar p $PACKAGE*.deb data.tar.xz | unxz | tar x -C /tiny
    elif [[ "$(ls data.*)" == "data.tar.zst" ]]; then
      ar p $PACKAGE*.deb data.tar.zst | unzstd | tar x -C /tiny
    else
      echo "Unsupported data format"
      exit 1
    fi
    source_package="$(dpkg-deb --showformat='${source:Package}' -W "$PACKAGE"*.deb)"
    source_version="$(dpkg-deb --showformat='${source:Version}' -W "$PACKAGE"*.deb)"
    source_upstream_version="$(dpkg-deb --showformat='${source:Upstream-Version}' -W "$PACKAGE"*.deb)"
    if [[ "$(ls control.*)" == "control.tar.xz" ]]; then
        unxz < control.tar.xz | tar x ./control && mv ./control "/tiny/var/lib/dpkg/status.d/$PACKAGE"
        echo "Source-Package: ${source_package}" >> "/tiny/var/lib/dpkg/status.d/$PACKAGE"
        echo "Source-Version: ${source_version}" >> "/tiny/var/lib/dpkg/status.d/$PACKAGE"
        echo "Source-Upstream-Version: ${source_upstream_version}" >> "/tiny/var/lib/dpkg/status.d/$PACKAGE"
    elif [[ "$(ls control.*)" == "control.tar.gz" ]]; then
        tar xzf control.tar.gz ./control && mv ./control "/tiny/var/lib/dpkg/status.d/$PACKAGE"
        echo "Source-Package: ${source_package}" >> "/tiny/var/lib/dpkg/status.d/$PACKAGE"
        echo "Source-Version: ${source_version}" >> "/tiny/var/lib/dpkg/status.d/$PACKAGE"
        echo "Source-Upstream-Version: ${source_upstream_version}" >> "/tiny/var/lib/dpkg/status.d/$PACKAGE"
    elif [[ "$(ls control.*)" == "control.tar.zst" ]]; then
        tar xzf control.tar.zst ./control && mv ./control "/tiny/var/lib/dpkg/status.d/$PACKAGE"
        echo "Source-Package: ${source_package}" >> "/tiny/var/lib/dpkg/status.d/$PACKAGE"
        echo "Source-Version: ${source_version}" >> "/tiny/var/lib/dpkg/status.d/$PACKAGE"
        echo "Source-Upstream-Version: ${source_upstream_version}" >> "/tiny/var/lib/dpkg/status.d/$PACKAGE"
    else
      echo "Unsupported control format"
        exit 1
    fi
    rm debian-binary data.tar.* control.tar.*
done < packagelist
