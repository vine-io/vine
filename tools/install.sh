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

version=`curl -s https://api.github.com/repos/vine-io/vine/releases/latest | grep tag_name | cut -d'"' -f4`
pname=vine_${version:1}_${os}_${archi}.tar.gz
package=`curl -s https://api.github.com/repos/vine-io/vine/releases/latest | grep browser_download_url | grep ${os} | cut -d'"' -f4 | grep "${pname}"`

echo "install package: ${package}"
wget ${package} -O /tmp/${pname} && tar -xvf /tmp/${pname} -C $gopath/bin

rm -fr /tmp/$pname
