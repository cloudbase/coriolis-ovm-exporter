apk add musl-dev gcc db-dev libressl-dev git libtool m4 autoconf g++ make

git clone --depth 1 https://salsa.debian.org/debian/db5.3.git

cd db5.3
cd dist && libtoolize -cfi && cd -
cd lang/sql/sqlite && libtoolize -cfi && cd -
cd dist && ./s_config
cd ../

sed -i 's/__atomic_compare_exchange/__atomic_compare_exchange_db/g' ./src/dbinc/atomic.h

cd build_unix/


../dist/configure --prefix=/usr --enable-cxx --enable-dbm --enable-compat185 --enable-stl --enable-static

make -j8
make install

export GO111MODULE=auto

go build -ldflags="-extldflags=-static"
