var express    = require('express');
var bodyParser = require('body-parser');

var app = express();
app.use(bodyParser.urlencoded({ extended: true }));
app.use(bodyParser.json());

var router = express.Router();
router.post('/echo', function(req, res) {
  res.json({ msg: req.body.msg });
});

app.use('/api', router);
app.listen(8080);
