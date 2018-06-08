#!/bin/bash

TARGET=apiserver
TARGET_TAR=apiserver.tar.gz
TARGET_PATH=$GOPATH/src/github.com/fabtestorg/test_fabric/$TARGET/


echo "----------build apiserver image----------"

if [ -f ./$TARGET_TAR ]; then
    rm ./$TARGET_TAR
fi
#rm ./apiserver.tar.gz

if [ -d ./$TARGET ]; then
    echo "remove old apiserver file"
    rm -rf ./$TARGET
fi
mkdir ./$TARGET
#rm -rf ./apiserver
#mkdir apiserver


if [ "$1" != "0" ]; then
    echo "build apiserver wait ...."
    cd $GOPATH/src/github.com/fabtestorg/test_fabric/apiserver
    go build --ldflags "-extldflags -static"
    cd -
#    cp $GOPATH/src/github.com/peersafe/factoring/apiserver/apiserver ./apiserver/
fi

if [ -f $TARGET_PATH/$TARGET ]; then
    cp $TARGET_PATH/$TARGET ./$TARGET/
else
    echo "--------ERROR: apiserver process has not been build.----------"
    exit
fi

#if [ -f $TARGET_PATH/sdk/client_sdk.yaml ]; then
#    cp $TARGET_PATH/sdk/client_sdk.yaml .
#else
#    echo "----------ERROR: client_sdk.yaml does not exist----------"
#    exit
#fi

if [ -d $TARGET_PATH/schema ]; then
    cp -r $TARGET_PATH/schema ./$TARGET/
else
    echo "----------ERROR: schema does not exist----------"
    exit
fi

tar -zcvf $TARGET_TAR $TARGET

image_id=`docker images |grep "apiserver"|awk '{print $3}'`
if  [ $image_id -ne 0 ]; then
    echo "remove the apiserver image $image_id"
    docker rmi $image_id
fi

docker build -t test_fabric/$TARGET .


