Running compilation locally:

1. Build the build environment:
```shell
docker build --tag compilation-<target> --file <target>.Dockerfile .

# Jammy example
docker build --tag compilation-jammy --file jammy.Dockerfile .

# Noble example
docker build --tag compilation-noble --file noble.Dockerfile .
```

2. Make the output directory:
```shell
export output_dir=$(mktemp -d)
```

3. Run compilation and use a volume mount to access it:

When --os and --arch are omitted, --os defaults to `linux` and --arch defaults to `amd64` for backward compatibility.

```shell
docker run --volume $output_dir:/tmp/compilation compilation-<target> --outputDir /tmp/compilation --target <target> --version <version> --os <os> --arch <arch>

# Jammy example
docker run --volume $output_dir:/tmp/compilation compilation-jammy --outputDir /tmp/compilation --target jammy --version 72.1

# Noble example
docker run --volume $output_dir:/tmp/compilation compilation-noble --outputDir /tmp/compilation --target noble --version 72.1
```
