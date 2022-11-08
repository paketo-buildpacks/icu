Running compilation locally:

1. Build the build environment:
```shell
docker build --tag compilation-<target> --file <target>.Dockerfile .

# Bionic example
docker build --tag compilation-bionic --file bionic.Dockerfile .

# Jammy example
docker build --tag compilation-jammy --file jammy.Dockerfile .
```

2. Make the output directory:
```shell
export output_dir=$(mktemp -d)
```

3. Run compilation and use a volume mount to access it:
```shell
docker run --volume $output_dir:/tmp/compilation compilation-<target> --outputDir /tmp/compilation --target <target> --version <version> 

# Bionic example
docker run --volume $output_dir:/tmp/compilation compilation-bionic --outputDir /tmp/compilation --target bionic --version 72.1

# Jammy example
docker run --volume $output_dir:/tmp/compilation compilation-jammy --outputDir /tmp/compilation --target jammy --version 72.1
```
