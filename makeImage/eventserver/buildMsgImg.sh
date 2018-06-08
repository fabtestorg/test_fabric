#!/bin/bash

TARGET=eventserver
TARGET_TAR=eventserver.tar.gz
TARGET_PATH=$GOPATH/src/github.com/fabtestorg/test_fabric/$TARGET/


echo "----------build eventserver image----------"

if [ -f ./$TARGET_TAR ]; then
    rm ./$TARGET_TAR
fi
#rm ./eventserver.tar.gz

if [ -d ./$TARGET ]; then
    echo "remove old eventserver file"
    rm -rf ./$TARGET
fi
mkdir ./$TARGET
#rm -rf ./eventserver
#mkdir eventserver


if [ "$1" != "0" ]; then
    echo "build eventserver wait ...."
    cd $GOPATH/src/github.com/fabtestorg/test_fabric/eventserver
    go build --ldflags "-extldflags -static"
    cd -
#    cp $GOPATH/src/github.com/peersafe/factoring/eventserver/eventserver ./eventserver/
fi

if [ -f $TARGET_PATH/$TARGET ]; then
    cp $TARGET_PATH/$TARGET ./$TARGET/
else
    echo "--------ERROR: eventserver process has not been build.----------"
    exit
fi

#if [ -f $TARGET_PATH/client_sdk.yaml ]; then
#    cp $TARGET_PATH/client_sdk.yaml .
#else
#    echo "----------ERROR: client_sdk.yaml does not exist----------"
#    exit
#fi


tar -zcvf $TARGET_TAR ./$TARGET

image_id=`docker images |grep "eventserver"|awk '{print $3}'`
if  [ $image_id -ne 0 ]; then
    echo "remove the eventserver image $image_id"
    docker rmi -f $image_id
fi

#docker rmi factoring/eventserver:latest
docker build -t test_fabric/$TARGET .
