#!/usr/bin/env bash

set -eu
set -o pipefail

readonly PROG_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly STACK_DIR="$(cd "${PROG_DIR}/.." && pwd)"
readonly BIN_DIR="${STACK_DIR}/.bin"
readonly BUILD_DIR="${STACK_DIR}/build"

# shellcheck source=SCRIPTDIR/.util/tools.sh
source "${PROG_DIR}/.util/tools.sh"

# shellcheck source=SCRIPTDIR/.util/print.sh
source "${PROG_DIR}/.util/print.sh"

function main() {
  local build run receiptFilename buildReceipt runReceipt receipts
  build="${BUILD_DIR}/build.oci"
  run="${BUILD_DIR}/run.oci"
  receiptFilename="receipt.cyclonedx.json"
  buildReceipt="${BUILD_DIR}/build-${receiptFilename}"
  runReceipt="${BUILD_DIR}/run-${receiptFilename}"

  while [[ "${#}" != 0 ]]; do
    case "${1}" in
      --help|-h)
        shift 1
        usage
        exit 0
        ;;

      --build-image|-b)
        build="${2}"
        shift 2
        ;;

      --run-image|-r)
        run="${2}"
        shift 2
        ;;

      --build-receipt|-B)
        buildReceipt="${2}"
        shift 2
        ;;

      --run-receipt|-R)
        runReceipt="${2}"
        shift 2
        ;;

      "")
        # skip if the argument is empty
        shift 1
        ;;

      *)
        util::print::error "unknown argument \"${1}\""
    esac
  done

  tools::install

		
  # We are generating receipts for all platforms
  receipts::generate::multi::arch "${build}" "${run}" "${buildReceipt}" "${runReceipt}"

  util::print::success "Success! Receipts are:\n  ${buildReceipt}\n  ${runReceipt}\n"
}

function usage() {
  cat <<-USAGE
receipts.sh [OPTIONS]

Generates receipts listing packages installed on build and run images of the
stack.

OPTIONS
  --help          -h  prints the command usage
  --build-image   -b  path to OCI image of build image. Defaults to
                      ${BUILD_DIR}/build.oci
  --run-image     -r  path to OCI image of build image
                      ${BUILD_DIR}/run.oci
  --build-receipt -B  path to output build image package receipt. Defaults to
                      ${BUILD_DIR}/build-receipt.cyclonedx.json
  --run-receipt   -R  path to output run image package receipt. Defaults to
                      ${BUILD_DIR}/run-receipt.cyclonedx.json
USAGE
}

function tools::install() {
  util::tools::crane::install \
    --directory "${BIN_DIR}"
  util::tools::jam::install \
    --directory "${BIN_DIR}"
  util::tools::syft::install \
    --directory "${BIN_DIR}"
}

# Generates syft receipts for each architecture for given oci archives
function receipts::generate::multi::arch() {
  local buildArchive runArchive registryPort registryPid localRegistry imageType archiveName imageReceipt

  buildArchive="${1}"
  runArchive="${2}"
  buildOutput="${3}"
  runOutput="${4}"
  
  registryPort=$(get::random::port)
  registryPid=$(local::registry::start $registryPort)
  localRegistry="127.0.0.1:$registryPort"

  # Push the oci archives to the local registry
  jam publish-stack \
    --build-ref "$localRegistry/build" \
    --build-archive $buildArchive \
    --run-ref "$localRegistry/run" \
    --run-archive $runArchive

  # Ensure we can write to the BUILD_DIR
  if [ $(stat -c %u build) = "0" ]; then
    sudo chown -R "$(id -u):$(id -g)" "$BUILD_DIR"
  fi

  for archivePath in "${buildArchive}" "${runArchive}" ; do
    archiveName=$(basename "${archivePath}")        # either 'build.oci' or 'run.oci'
    imageType=$(basename -s .oci "${archivePath}")  # either 'build' or 'run'

    util::print::title "Generating package SBOM for ${archiveName}"

    for imageArch in $(crane manifest "$localRegistry/$imageType" | jq -r '.manifests[].platform.architecture'); do
      if [[ "$imageType" = "build" ]]; then
        dir=$(dirname ${buildOutput})
        fileName=$(basename ${buildOutput})
        imageReceipt="${dir}/${imageArch}-${fileName}"
      elif [[ "$imageType" = "run" ]]; then
        dir=$(dirname ${runOutput})
        fileName=$(basename ${runOutput})
        imageReceipt="${dir}/${imageArch}-${fileName}"
      fi

      util::print::info "Generating CycloneDX package SBOM using syft for $archiveName on platform linux/$imageArch saved as $imageReceipt"

      # Generate the architecture-specific SBOM from image in the local registry
      syft packages "registry:$localRegistry/$imageType" \
        --output cyclonedx-json \
        --file "$imageReceipt" \
        --platform "linux/$imageArch"
    done
  done

  kill $registryPid
}

# Returns a random unused port
function get::random::port() {
  local port=$(shuf -i 50000-65000 -n 1)
  netstat -lat | grep $port > /dev/null
  if [[ $? == 1 ]] ; then
    echo $port
  else
    echo get::random::port
  fi
}

# Starts a local registry on the given port and returns the pid
function local::registry::start() {
  local registryPort registryPid localRegistry

  registryPort="$1"
  localRegistry="127.0.0.1:$registryPort"

  # Start a local in-memory registry so we can work with oci archives
  PORT=$registryPort crane registry serve --insecure > /dev/null 2>&1 &
  registryPid=$!

  # Stop the registry if execution is interrupted
  trap "kill $registryPid" 1 2 3 6

  # Wait for the registry to be available
  until crane catalog $localRegistry > /dev/null 2>&1; do
    sleep 1
  done

  echo $registryPid
}

main "${@:-}"
