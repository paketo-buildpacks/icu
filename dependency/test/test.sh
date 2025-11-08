#!/usr/bin/env bash

set -euo pipefail
shopt -s inherit_errexit

main() {
  local tarball_path version os arch

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

      --os)
        os="${2}"
        shift 2
        ;;

      --arch)
        arch="${2}"
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

  # When --os and --arch are provided, the --platform arg is passed to docker build and run commands.
  # This assumes the runner has qemu and buildkit set up, and that the docker daemon and cli experimental features are enabled.
  docker_platform_arg=""
  if [[ "${os}" != "" && "${arch}" != "" ]]; then
    docker_platform_arg="--platform ${os}/${arch}"
    echo "docker commands will be called with ${docker_platform_arg}"
  fi

  if [[ ${artifact} == *"noble"* ]]; then
    echo "Running noble test..."
    docker build -t test-noble -f noble.Dockerfile ${docker_platform_arg} .
    docker run --rm -v "${dir}:/input" ${docker_platform_arg} test-noble --tarballPath "/input/${artifact}" --expectedVersion "${version}" --os "${os}" --arch "${arch}"

  elif [[ ${artifact} == *"jammy"* ]]; then
    echo "Running jammy test..."
    docker build -t test-jammy -f jammy.Dockerfile ${docker_platform_arg} .
    docker run --rm -v "${dir}:/input" ${docker_platform_arg} test-jammy --tarballPath "/input/${artifact}" --expectedVersion "${version}" --os "${os}" --arch "${arch}"
  else
    echo "noble or jammy not found - skipping tests"
  fi
}

main "${@:-}"
