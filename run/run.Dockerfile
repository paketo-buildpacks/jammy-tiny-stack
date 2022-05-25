FROM ubuntu:jammy AS builder

ARG packages

RUN apt-get update && \
  apt-get install -y xz-utils binutils zstd

ADD install-certs.sh .

ADD files/passwd /tiny/etc/passwd
ADD files/nsswitch.conf /tiny/etc/nsswitch.conf
ADD files/group /tiny/etc/group

RUN mkdir -p /tiny/tmp /tiny/var/lib/dpkg/status.d/

# We can't use dpkg -i (even with --instdir=/tiny) because we don't want to
# install the dependencies, and dpkg-deb has no way to ignore all dependencies;
# each dependency must be explicitly listed
RUN apt download $packages \
    && for pkg in $packages; do \
      dpkg-deb --field $pkg*.deb > /tiny/var/lib/dpkg/status.d/$pkg \
      && dpkg-deb --extract $pkg*.deb /tiny; \
    done

RUN ./install-certs.sh

RUN find /tiny/usr/share/doc/*/* ! -name copyright | xargs rm -rf && \
  rm -rf \
    /tiny/etc/update-motd.d/* \
    /tiny/usr/share/man/* \
    /tiny/usr/share/lintian/*

ADD files/os-release /tmp/etc/os-release

RUN grep -v 'PRETTY_NAME=' "/tiny/etc/os-release" \
      | grep -v 'HOME_URL=' \
      | grep -v 'SUPPORT_URL=' \
      | grep -v 'BUG_REPORT_URL=' \
      | tee /tmp/etc/os-release-upstream \
    && rm /tiny/etc/os-release \
    && cat /tmp/etc/os-release-upstream /tmp/etc/os-release \
      | tee /tiny/etc/os-release

# Distroless images use /var/lib/dpkg/status.d/<file> instead of /var/lib/dpkg/status
# TODO: can we rm the file entirely?
RUN echo "" > /tiny/var/lib/dpkg/status

FROM scratch
COPY --from=builder /tiny/ /

