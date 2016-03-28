# checkout the code
git clone https://github.com/google/protobuf.git && cd protobuf

# current latest release is v3b2
git checkout v3.0.0-beta-2

# config and install
./autogen.sh && ./configure
make && sudo make install

# install the go protoc plugin
github.com/golang/protobuf/protoc-gen-go
