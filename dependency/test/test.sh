#!/usr/bin/env bash

set -euo pipefail
shopt -s inherit_errexit

main() {
  local tarball_path version

  while [ "${#}" != 0 ]; do
    case "${1}" in
      --expectedVersion)
        version="${2}"
        shift 2
        ;;

      --tarballPath)
        tarball_path="${2}"
        shift 2
        ;;

      "")
        shift
        ;;

      *)
        echo "unknown argument \"${1}\""
        exit 1
    esac
  done

  if [[ -z "${version:-}" || -z "${tarball_path:-}" ]]; then
    echo "version and tarballPath are required required"
    exit 1
  fi

  dir="$(dirname "${tarball_path}")"
  artifact="$(basename "${tarball_path}")"

  if [[ ${artifact} == *"bionic"* ]]; then
    echo "Running bionic test..."
    docker build -t test-bionic -f bionic.Dockerfile .
    docker run --rm -v "${dir}:/input" test-bionic --tarballPath "/input/${artifact}" --expectedVersion "${version}"

  elif [[ ${artifact} == *"jammy"* ]]; then
    echo "Running jammy test..."
    docker build -t test-jammy -f jammy.Dockerfile .
    docker run --rm -v "${dir}:/input" test-jammy --tarballPath "/input/${artifact}" --expectedVersion "${version}"
  else
    echo "bionic or jammy not found - skipping tests"
  fi
}

main "${@:-}"
