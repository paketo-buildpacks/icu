To test locally:

```shell
# assume $output_dir is the output from the compilation step, with a tarball and a checksum in it

docker run -it \
  --volume $output_dir:/tmp/output_dir \
  --volume $PWD:/tmp/test \
  ubuntu:jammy \
  bash

# Now on the container

# Passing
$ /tmp/test/test.sh \
  --tarballPath /tmp/output_dir/icu_71.1_linux_jammy_1b1ca43f.tgz \
  --expectedVersion 71.1
tarballPath=/tmp/output_dir/icu_71.1_linux_jammy_1b1ca43f.tgz
expectedVersion=71.1
All tests passed!

# Failing
$ /tmp/test/test.sh \
  --tarballPath /tmp/output_dir/icu_71.1_linux_jammy_1b1ca43f.tgz \
  --expectedVersion 71.0
tarballPath=/tmp/output_dir/icu_71.1_linux_jammy_c3a27edf.tgz
expectedVersion=71.0
Version 71.1 does not match expected version 71.0
```
