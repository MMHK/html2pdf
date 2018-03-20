const puppeteer = require('puppeteer');

let output,
    url;
    
if (system.args.length > 2) {
    url = system.args[1];
    output = system.args[2];
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