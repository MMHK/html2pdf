(async function() {

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
    const viewPortWidth = document.querySelector(".page").clientWidth;
    const styleSheets = document.querySelectorAll('style');
    for (const sheet of styleSheets) {
        sheet.innerHTML = sheet.innerHTML.replace(/([0-9\.]+)px/g, (match, p1) => {
            return `${(p1 / viewPortWidth * 100) }vw`;
        });
    }

    // add page print style
    const printStyle = document.createElement('style');
    printStyle.setAttribute("type", "text/css");
    printStyle.innerHTML = `
        @page {
			size: A4 portrait;
			margin: 0em;
		}
		@media print {
            .page {
                padding: 0;
                height: 99vh;
                page-break-before: always;
            }
            .page:first-child {
                page-break-before: avoid;
            }
		}
    `;
    document.head.appendChild(printStyle);
})();
