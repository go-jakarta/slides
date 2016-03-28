#!/usr/bin/env node

var grpc = require('grpc');

// load service definition
var PROTO_PATH = __dirname + '/../hello.proto';
var helloworld = grpc.load(PROTO_PATH).helloworld;

// sayHello says hi.
function sayHello(req, cb) {
  // create response
  var res = { reply: 'hello ' + req.request.greeting + ' from nodejs' };

  // send response
  cb(null, res);
}

(function() {
  // create service definition
  var server = new grpc.Server();
  server.addProtoService(helloworld.HelloService.service, { sayHello: sayHello });

  // create listener
  server.bind('0.0.0.0:8833', grpc.ServerCredentials.createInsecure());

  // run server
  server.start();
})();
