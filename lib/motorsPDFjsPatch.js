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
        const styleSheets = document.querySelectorAll('style');
        for (const sheet of styleSheets) {
            sheet.innerHTML = sheet.innerHTML.replace(/([:\ ]+)([0-9\.\-]+)px([;\ }]+)/g, (match, p1, p2, p3) => {
                return `${p1}${(p2 / viewPortWidth * 100) }vw${p3}`;
            });
        }
        const body = document.body;
        body.innerHTML = body.innerHTML.replace(/([:\ ]+)([0-9\.\-]+)px([;\ }]+)/g, (match, p1, p2, p3) => {
            return `${p1}${(p2 / viewPortWidth * 100) }vw${p3}`;
        });

        // add page print style
        const printStyle = document.createElement('style');
        const bodyFontSizeVW = `${bodyFontSize/viewPortWidth*100}vw`;
        printStyle.setAttribute("type", "text/css");
        printStyle.innerHTML = `
        @page {
			size: A4 portrait;
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
            transform: translate(-50%, -50%) rotate(-45deg);;
        }
		@media print {
		    body {
                font-size: ${bodyFontSizeVW};
            }
            .page, body > div {
                padding: 0;
                height: 99vh;
                position: relative;
                background-size: 100% auto;
                page-break-before: always;
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