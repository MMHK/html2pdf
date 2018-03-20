const puppeteer = require('puppeteer');

let output,
    system = require("process"),
    url;
    
if (system.argv.length > 3) {
    url = system.argv[2];
    output = system.argv[3];
}

(async() => {

    const browser = await puppeteer.launch({args: [
        '--no-sandbox',
        '--disable-setuid-sandbox'
    ]});
    
    const page = await browser.newPage();
    await page.goto(url, {waitUntil: 'networkidle2'});
    await page.emulateMedia('screen');
    await page.pdf({
        path: output, 
        printBackground: true,
        margin: {
            top: "0cm",
            right: "0cm",
            bottom: "0cm",
            left: "0cm",
        },
        width: "1240px",
        height: "1754px"
    });

    await browser.close();
})();
