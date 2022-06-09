# Paketo Jammy Tiny Stack

## What is a stack?
See Paketo's [stacks documentation](https://paketo.io/docs/concepts/stacks/).

## What is this stack for?
Ideal for:
- most Golang apps
- Java [GraalVM Native Images](https://www.graalvm.org/docs/reference-manual/native-image/)

## What's in the build and run images of this stack?
This stack's build image is based on Ubuntu Jammy Jellyfish. Its run image does not include a Linux distribution.

- To see the **list of all packages installed** in the build or run image for a given release,
see the `jammy-tiny-stack-{version}-build-receipt.txt` and 
`jammy-tiny-stack-{version}-run-receipt.txt` attached to each
[release](https://github.com/paketo-buildpacks/jammy-tiny-stack/releases). For a quick overview
of the packages you can expect to find, see the [stack descriptor file](stack/stack.toml).

- To generate a package receipt based on existing `build.oci` and `run.oci` archives, use [`scripts/receipts.sh`](scripts/receipts.sh).

## How can I contribute?
Contribute changes to this stack via a Pull Request. Depending on the proposed changes,
you may need to [submit an RFC](https://github.com/paketo-buildpacks/rfcs) first.

### How do I test the stack locally?
Run [`scripts/test.sh`](scripts/test.sh).
