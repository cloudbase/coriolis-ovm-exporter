FROM golang:alpine

WORKDIR /root

RUN apk add musl-dev gcc db-dev libressl-dev git libtool m4 autoconf g++ make
RUN git clone --depth 1 https://salsa.debian.org/debian/db5.3.git

WORKDIR /root/db5.3

RUN cd dist &&  libtoolize -cfi && cd .. && \
    cd lang/sql/sqlite && libtoolize -cfi && cd - && \
    cd dist && ./s_config

RUN sed -i 's/__atomic_compare_exchange/__atomic_compare_exchange_db/g' src/dbinc/atomic.h

RUN cd build_unix/ && ../dist/configure --prefix=/usr \
    --enable-cxx --enable-dbm --enable-compat185 \
    --enable-stl --enable-static && make -j8 && make install

ADD ./scripts/build-static.sh /build-static.sh
RUN chmod +x /build-static.sh

CMD ["/bin/sh"]