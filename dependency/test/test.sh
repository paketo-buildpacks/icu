

#!/bin/bash

set -euo pipefail

extract_tarball() {
  rm -rf icu
  mkdir icu
  tar --extract \
    --file "$1" \
    --directory icu
}

set_ld_library_path() {
  export LD_LIBRARY_PATH="$PWD/icu/lib:${LD_LIBRARY_PATH:-}"
}

check_version() {
  expected_version=$1
  actual_version="$(/icu/bin/icuinfo 2> /dev/null | sed -rn 's/.*param name="version">([0-9\.]+).*/\1/p')"
  if [[ "${actual_version}" != "${expected_version}" ]]; then
    echo "Version ${actual_version} does not match expected version ${expected_version}"
    exit 1
  fi
}

check_file() {
  if ! test -f icu/lib/libicudata.so; then
    echo "Library file missing"
    exit 1
  fi
}

main() {
  local tarballPath expectedVersion
  tarballPath=""
  expectedVersion=""

  while [ "${#}" != 0 ]; do
    case "${1}" in
      --tarballPath)
        tarballPath="${2}"
        shift 2
        ;;

      --expectedVersion)
        expectedVersion="${2}"
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

  if [[ "${tarballPath}" == "" ]]; then
    echo "--tarballPath is required"
    exit 1
  fi

  if [[ "${expectedVersion}" == "" ]]; then
    echo "--expectedVersion is required"
    exit 1
  fi

  echo "tarballPath=${tarballPath}"
  echo "expectedVersion=${expectedVersion}"

  extract_tarball "${tarballPath}"
  set_ld_library_path
  check_version "${expectedVersion}"
  check_file

  echo "All tests passed!"
}

main "$@"
