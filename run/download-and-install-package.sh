#!/bin/bash

# TODO: Why do we install packages this way instead of apt-get installing them?
# TODO: can we use dpkg --instdir=/tiny/ -i $PACKAGE*.deb?
# TODO: dpkg-deb --verbose --raw-extract whatever.deb /some/path
set -eu -o pipefail

mkdir -p "/tiny/var/lib/dpkg/status.d"
apt download $(cat packagelist | tr '\n' ' ')

while read PACKAGE; do
    echo "installing $PACKAGE..."

    package_info_file=/tiny/var/lib/dpkg/status.d/$PACKAGE

    source_package="$(dpkg-deb --showformat='${source:Package}' -W "$PACKAGE"*.deb)"
    source_version="$(dpkg-deb --showformat='${source:Version}' -W "$PACKAGE"*.deb)"
    source_upstream_version="$(dpkg-deb --showformat='${source:Upstream-Version}' -W "$PACKAGE"*.deb)"

    dpkg-deb -f $PACKAGE*.deb > "$package_info_file"
    echo "Source-Package: ${source_package}" >> "$package_info_file"
    echo "Source-Version: ${source_version}" >> "$package_info_file"
    echo "Source-Upstream-Version: ${source_upstream_version}" >> "$package_info_file"

    #dpkg-deb --verbose --raw-extract $PACKAGE*.deb ./$package_tmp
    # dpkg-deb --raw-extract $PACKAGE*.deb ./tiny
    # rm -rf /tiny/DEBIAN

    # We can't use dpkg -i (even with --instdir) because we don't want to satisfy the dependencies.
    # E.g. base-files depends on awk but we don't want awk on the resuling image.
    #dpkg --instdir=/tiny/ -i $PACKAGE*deb

    package_tmp="$PACKAGE"_tmp
    mkdir -p "$package_tmp"
    pushd "$package_tmp" > /dev/null
      ar x ../$PACKAGE*.deb

      if [[ "$(ls data.*)" == "data.tar.xz" ]]; then
        ar p ../$PACKAGE*.deb data.tar.xz | unxz | tar x -C /tiny
      elif [[ "$(ls data.*)" == "data.tar.zst" ]]; then
        ar p ../$PACKAGE*.deb data.tar.zst | unzstd | tar x -C /tiny
      else
        echo "Unsupported data format"
        exit 1
      fi
    popd > /dev/null

    rm -rf "$package_tmp"
done < packagelist
