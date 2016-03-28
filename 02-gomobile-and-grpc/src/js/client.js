#!/usr/bin/env node

var grpc = require('grpc');

// load service definition
var PROTO_PATH = __dirname + '/../hello.proto';
var helloworld = grpc.load(PROTO_PATH).helloworld;

(function() {
  // create client
  var client = new helloworld.HelloService('localhost:8833', grpc.credentials.createInsecure());

  // create request
  var req = { greeting: 'ken' };

  // call method
  client.sayHello(req, function(err, res) {
    if (err) {
      console.log('error: ' + err);
      return
    }

    // handle response
    console.log('Received: "' + res.reply + '"');
  });
})();
