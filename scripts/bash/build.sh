#!/bin/bash
bin=$(basename $(pwd))
os=linux
archs=(amd64 arm arm64)
build_root=build/
bin_path=$build_root/$bin
# version=""
container=0

for arg in "$@"
do
  case $arg in
    # --version)
    #   shift
    #   version=$1
    #   shift
    #   ;;
    --image)
      container=1
      shift
      ;;
    --cmd)
      shift
      archs=(amd64)
      shift
      ;;
  esac
done

if [ -z $bin ]; then
  bin=$(echo $(basename $(pwd)))
fi

# if [ -z $version ]; then
#   echo "A version must be specified."
#   exit 1
# fi

if [ -z $build_root ]; then
  exit 1
fi

if [ -d $build_root ]; then
  rm -rf $build_root/*
fi

mkdir -p $build_root

go test ./...
if [ $? -ne 0 ]; then
  exit 1
fi

for arch in ${archs[@]}
do
  echo "Building $bin for $os/$arch"
  CGO_ENABLED=0 GOOS=$os GOARCH=$arch go build \
    -o ${bin_path}_${arch} \
    -ldflags="-s -w" \
    -trimpath main.go

  if [ $container -eq 1 ] && [ "$os" == "linux" ]; then
    echo "Building container $bin:$version for $os/$arch"
    docker build -t $bin:$version --platform $os/$arch .
  fi
done