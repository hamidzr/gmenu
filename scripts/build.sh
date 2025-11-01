#!/bin/sh

set -x

# Define the targets with both amd64 and arm64 architectures
targets=(
  "darwin/amd64"
  "darwin/arm64"
)

# Create the bin directory if it doesn't exist
mkdir -p bin

# Loop through each target and build the executable
for target in "${targets[@]}"; do
  os=${target%/*}
  arch=${target#*/}
  output_name="bin/gmenu-${os}-${arch}"
  [ "$os" = "windows" ] && output_name="${output_name}.exe"
  echo "Building for $os/$arch..."

  # Enable CGO and set the target OS and architecture
  env CGO_ENABLED=1 GOOS=$os GOARCH=$arch go build -o $output_name -v ./cmd
done

echo "Build complete."
