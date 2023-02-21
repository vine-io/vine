#!/usr/bin/env bash

archi=`arch`
if [ "$archi" == "x86_64" ];then
  archi="amd64"
elif [ "$archi" == "i386" ];then
  archi="arm64"
fi

os=`uname | tr '[A-Z]' '[a-z]'`

go_version=`go version`
if [ "$go_version" == "" ];then
    echo "missing golang binary"
    exit 1
fi

gopath=`go env var GOPATH | grep "/"`
if [ "$gopath" == "" ];then
  gopath=`go env var GOROOT | grep "/"`
fi

package=`curl -s https://api.github.com/repos/vine-io/vine/releases/latest | grep browser_download_url | grep ${os} | cut -d'"' -f4 | grep "vine-${os}-${archi}"`

echo "install package: ${package}"
wget ${package} -O /tmp/vine.tar.gz && tar -xvf /tmp/vine.tar.gz -C /tmp/

mv /tmp/${os}-${archi}/* $gopath/bin

rm -fr /tmp/${os}-${archi}
rm -fr /tmp/vine.tar.gz
