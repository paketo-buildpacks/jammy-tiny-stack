#!/bin/bash

# We can't use dpkg -i (even with --instdir=/tiny) because we don't want to
# install the dependencies, and dpkg-deb has no way to ignore all dependencies;
# each dependency must be explicitly listed

set -eu -o pipefail

apt download $(cat packagelist | tr '\n' ' ')

while read PACKAGE; do
    echo "installing $PACKAGE..."

    dpkg-deb -f $PACKAGE*.deb > /tiny/var/lib/dpkg/status.d/$PACKAGE

    dpkg-deb --extract $PACKAGE*.deb ./tiny
done < packagelist
