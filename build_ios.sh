#!/usr/bin/env bash
set -e

MODULE=demo

ROOT_PATH=$(pwd)
OUTPUT_PATH=$ROOT_PATH/output/ios/
if [ ! -d $OUTPUT_PATH ]; then mkdir -p $OUTPUT_PATH; fi
if [ ! -d $ROOT_PATH/openssl-ios/ ]; then
    curl -O https://getseacatiostoracc.blob.core.windows.net/getseacatio/openssl/openssl-dev-1.0.2o-ios.tar.gz
    tar xvf openssl-dev-1.0.2o-ios.tar.gz
    mv openssl openssl-ios
fi
OPENSSL_PATH="$ROOT_PATH/openssl-ios"
OPENSSL_INCLUDE="$OPENSSL_PATH/include/"
OPENSSL_LIBS="$OPENSSL_PATH/lib/"


export GOARCH=amd64
case "$OSTYPE" in
    solaris*) echo "SOLARIS" ;;
    darwin*)  export GOOS=darwin ;;
    linux*)   export GOOS=linux ;;
    bsd*)     echo "BSD" ;;
    msys*)    echo "WINDOWS" ;;
    *)        echo "unknown: $OSTYPE" ;;
esac

GOMOBILE=$ROOT_PATH/bin/gomobile
export GOPATH=$ROOT_PATH:$ROOT_PATH/app
PATH=$PATH:$ROOT_PATH/bin

MODULE_PATH=$ROOT_PATH/app/src/$MODULE
echo GOPATH: $GOPATH
echo MODULE_PATH: $MODULE_PATH
echo
cd $MODULE_PATH

# download and init gomobile if doesn't exists
GOMOBILE=$ROOT_PATH/bin/gomobile
if [ ! -e $GOMOBILE ]; then
    go get golang.org/x/mobile/cmd/gomobile
    $GOMOBILE init
fi

# download packages, but don't to build them
go get -t -d -v ./
gogetRetVal=$?
if [ $gogetRetVal -ne 0 ]; then
    cd $ROOT_PATH
    exit $gogetRetVal
fi
echo

cd $OUTPUT_PATH

export CGO_ENABLED=1
export GOGCCFLAGS="-I$OPENSSL_INCLUDE"
export CGO_CFLAGS="-I$OPENSSL_INCLUDE"
export CGO_CPPFLAGS="-I$OPENSSL_INCLUDE"
export CGO_CXXFLAGS="-I$OPENSSL_INCLUDE"
export CGO_FFLAGS=""
export CGO_LDFLAGS="-g -O2 -L$OPENSSL_LIBS -lssl -lcrypto"

# $GOMOBILE build -target=ios -ldflags="-L$OPENSSL_LIBS -lssl -lcrypto" -x $MODULE
$GOMOBILE build -target=ios -x $MODULE
goRetVal=$?
if [ $goRetVal -ne 0 ]; then
    cd $ROOT_PATH
    exit $goRetVal
fi
cd $ROOT_PATH
