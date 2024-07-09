(async function (){
    await (async function() {

        // Helper function to convert a blob to a Data URL
        async function blobToDataURL(blob) {
            return new Promise((resolve, reject) => {
                const reader = new FileReader();
                reader.onloadend = () => resolve(reader.result);
                reader.onerror = reject;
                reader.readAsDataURL(blob);
            });
        }

        // Helper function to fetch a resource and convert it to a Data URL
        async function fetchResourceAsDataURL(url) {
            const response = await fetch(url);
            const blob = await response.blob();
            return blobToDataURL(blob);
        }

        function convertRelativeURLs(baseURL, cssText) {
            return cssText.replace(/url\((?!['"]?(?:data|http|https):)['"]?([^'")]+)['"]?\)/g, (match, relativeURL) => {
                const absoluteURL = new URL(relativeURL, baseURL).href;
                return `url(${absoluteURL})`;
            });
        }

        function getBodyFontSize() {
            const body = document.body;
            const style = window.getComputedStyle(body);
            return style.fontSize.replace("px", "");
        }

        async function embedAssets() {
            const styleSheets = document.querySelectorAll('style');
            for (const sheet of styleSheets) {
                let urlList = [];
                let urlIndex = 0;
                // 先处理掉 url 数据
                sheet.innerHTML = sheet.innerHTML.replace(/url\(([^\)]+)\)/ig, (match, p1) => {
                    const url = p1.replace(/^['"]{0,1}(.*)['"]{0,1}$/, '$1')
                    if (/^data/i.test(url)) {
                        return match;
                    }
                    if (/(\.woff2|\.ttf|\.eot|\.svg#)/i.test(url)) {
                        return match;
                    }
                    urlList.push(url);
                    const raw = `{URL${urlIndex}}`;
                    urlIndex++;
                    return raw;
                });
                await Promise.all(urlList.map(async (url, index) => {
                    const dataURL = await fetchResourceAsDataURL(url);
                    sheet.innerHTML = sheet.innerHTML.replace(`{URL${index}}`, `url(${dataURL})`);
                }))
            }
        }

        const bodyFontSize = getBodyFontSize();

        // Process all <link> elements (CSS)
        const linkElements = document.querySelectorAll('link[rel="stylesheet"]');
        for (const link of linkElements) {
            const url = link.href;
            const dataURL = await fetchResourceAsDataURL(url);
            if (link.rel === 'stylesheet') {
                const styleElement = document.createElement('style');
                const cssText = await (await fetch(dataURL)).text();
                const adjustedCSSText = convertRelativeURLs(url, cssText);
                styleElement.setAttribute("data-url", url);
                styleElement.textContent = adjustedCSSText;
                document.head.appendChild(styleElement);
                link.remove();
            } else {
                link.href = dataURL;
            }
        }

        //replace px
        const viewPortWidth = document.querySelector(".page, body > div").clientWidth;
        const viewPortHeight = document.querySelector(".page, body > div").clientHeight;
        const isLandscape = viewPortWidth > viewPortHeight;
        const styleSheets = document.querySelectorAll('style');
        for (const sheet of styleSheets) {
            let urlData = [];
            let urlIndex = 0;

            // 先处理掉 dataURL 数据
            sheet.innerHTML = sheet.innerHTML.replace(/url\([^\)]+\)/ig, (match, p1) => {
                urlData.push(match);
                const raw = `{URLDATA${urlIndex}}`;
                urlIndex++;
                return raw;
            });

            // 将 px 单位  全部 转换成 vw
            sheet.innerHTML = sheet.innerHTML.replace(/([:\ ]+)?([0-9\.\-]+)px([;\ }]+)?/ig, (match, p1, p2, p3) => {
                if (!p1) {
                    p1 = ''
                }
                if (!p3) {
                    p3 = ''
                }
                return `${p1}${(p2 / viewPortWidth * 100) }vw${p3}`;
            });

            // 将 dataURL 数据替换回来
            urlData.forEach((url, index) => {
                sheet.innerHTML = sheet.innerHTML.replace(`{URLDATA${index}}`, url);
            })
            sheet.innerHTML = sheet.innerHTML.replace('x-small', '1em');

        }
        const body = document.body;
        let urlData = [];
        let urlIndex = 0;

        // 先处理掉 dataURL 数据
        body.innerHTML = body.innerHTML.replace(/"data:[^"]+"/ig, (match, p1) => {
            urlData.push(match);
            const raw = `"{URLDATA${urlIndex}}"`;
            urlIndex++;
            return raw;
        });
        body.innerHTML = body.innerHTML.replace(/([:\ ]+)?([0-9\.\-]+)px([;\ }]+)?/ig, (match, p1, p2, p3) => {
            return `${p1}${(p2 / viewPortWidth * 100) }vw${p3}`;
        });
        // 将 dataURL 数据替换回来
        urlData.forEach((url, index) => {
            body.innerHTML = body.innerHTML.replace(`"{URLDATA${index}}"`, url);
        });

        await embedAssets();

        // add page print style
        const printStyle = document.createElement('style');
        const bodyFontSizeVW = `${bodyFontSize/viewPortWidth*100}vw`;
        printStyle.setAttribute("type", "text/css");
        printStyle.innerHTML = `
        @page {
			size: A4 ${isLandscape ? 'landscape' : 'portrait' };
			margin: 0em;
		}
        .SAMPLE:after {
            display: inline-block;
            position: absolute;
            content: "SAMPLE";
            opacity: 0.5;
            top: 50%;
            left: 50%;
            font-size: 20vw;            
            margin: 0;
            transform: translate(-50%, -50%) rotate(-45deg);
        }
		@media print {
		    body {
                font-size: ${bodyFontSizeVW};
            }
            .page, body > div {
                padding: 0;
                height: 99.9vh;
                position: relative;
                background-size: 100% auto;
                page-break-before: always;
                transform: rotate(0deg);
            }
            .page:before, body > div:before {
                height: 100%;
            }
            .page:first-child, , body > div:first-child {
                page-break-before: avoid;
            }
            *[data-scrip-done=true] {
                display: none;
            }
		}
    `;
        document.head.appendChild(printStyle);
    })();

    const flag = document.createElement("span");
    flag.setAttribute("data-scrip-done", "true");
    document.body.appendChild(flag);
})()